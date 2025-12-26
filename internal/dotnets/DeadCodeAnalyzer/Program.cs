using Microsoft.Build.Locator;
using Microsoft.CodeAnalysis;
using Microsoft.CodeAnalysis.CSharp;
using Microsoft.CodeAnalysis.CSharp.Syntax;
using Microsoft.CodeAnalysis.MSBuild;
using System.Text.Json;

namespace DeadCodeAnalyzer;

class Program
{
    static async Task<int> Main(string[] args)
    {
        if (args.Length == 0)
        {
            Console.Error.WriteLine("用法: DeadCodeAnalyzer <项目文件.csproj 或 解决方案.sln>");
            return 1;
        }

        try
        {
            // 注册 MSBuild
            MSBuildLocator.RegisterDefaults();

            var analyzer = new Analyzer();
            var results = await analyzer.AnalyzeAsync(args[0]);

            // 输出 JSON 格式，方便 Go 解析
            var json = JsonSerializer.Serialize(results, new JsonSerializerOptions
            {
                WriteIndented = true
            });
            Console.WriteLine(json);

            return 0;
        }
        catch (Exception ex)
        {
            Console.Error.WriteLine($"错误: {ex.Message}");
            return 1;
        }
    }
}

class Analyzer
{
    // 定义：完整信息（包含位置）
    private readonly Dictionary<string, string> _defined = new();
    // 使用：简单标识
    private readonly HashSet<string> _used = new();

    public async Task<AnalysisResult> AnalyzeAsync(string projectPath)
    {
        var workspace = MSBuildWorkspace.Create();

        // 加载项目或解决方案
        if (projectPath.EndsWith(".sln"))
        {
            var solution = await workspace.OpenSolutionAsync(projectPath);

            // 分析解决方案中的所有项目
            foreach (var project in solution.Projects)
            {
                await AnalyzeProjectAsync(project);
            }
        }
        else
        {
            var project = await workspace.OpenProjectAsync(projectPath);
            await AnalyzeProjectAsync(project);
        }

        // 找出僵尸代码
        var deadCode = _defined.Where(kvp => !_used.Contains(kvp.Key))
                               .Select(kvp => kvp.Value)
                               .ToList();

        return new AnalysisResult
        {
            TotalDefined = _defined.Count,
            TotalUsed = _used.Count,
            DeadCodeItems = deadCode.Select(d => new DeadCodeItem
            {
                Symbol = d
            }).ToList()
        };
    }

    private async Task AnalyzeProjectAsync(Project project)
    {
        // 分析所有文档
        foreach (var document in project.Documents)
        {
            await AnalyzeDocumentAsync(document);
        }
    }

    private async Task AnalyzeDocumentAsync(Document document)
    {
        var syntaxTree = await document.GetSyntaxTreeAsync();
        var semanticModel = await document.GetSemanticModelAsync();

        if (syntaxTree == null || semanticModel == null)
            return;

        var root = await syntaxTree.GetRootAsync();

        // 收集定义
        CollectDefinitions(root, semanticModel, document.FilePath ?? "");

        // 收集使用
        CollectUsages(root, semanticModel);
    }

    private void CollectDefinitions(SyntaxNode root, SemanticModel model, string filePath)
    {
        // 1. 方法定义
        var methods = root.DescendantNodes().OfType<MethodDeclarationSyntax>();
        foreach (var method in methods)
        {
            var symbol = model.GetDeclaredSymbol(method);
            if (symbol != null && ShouldCheck(symbol))
            {
                var location = method.GetLocation().GetLineSpan();
                var key = $"{symbol.ContainingType?.Name}.{symbol.Name}";
                var value = $"{key}@{filePath}:{location.StartLinePosition.Line + 1}";
                _defined[key] = value;
            }
        }

        // 2. 属性定义
        var properties = root.DescendantNodes().OfType<PropertyDeclarationSyntax>();
        foreach (var prop in properties)
        {
            var symbol = model.GetDeclaredSymbol(prop);
            if (symbol != null && ShouldCheck(symbol))
            {
                var location = prop.GetLocation().GetLineSpan();
                var key = $"{symbol.ContainingType?.Name}.{symbol.Name}";
                var value = $"{key}@{filePath}:{location.StartLinePosition.Line + 1}";
                _defined[key] = value;
            }
        }

        // 3. 字段定义
        var fields = root.DescendantNodes().OfType<FieldDeclarationSyntax>();
        foreach (var field in fields)
        {
            foreach (var variable in field.Declaration.Variables)
            {
                var symbol = model.GetDeclaredSymbol(variable);
                if (symbol != null && ShouldCheck(symbol))
                {
                    var location = field.GetLocation().GetLineSpan();
                    var key = $"{symbol.ContainingType?.Name}.{symbol.Name}";
                    var value = $"{key}@{filePath}:{location.StartLinePosition.Line + 1}";
                    _defined[key] = value;
                }
            }
        }
    }

    private void CollectUsages(SyntaxNode root, SemanticModel model)
    {
        // 1. 查找所有标识符的使用
        var identifiers = root.DescendantNodes().OfType<IdentifierNameSyntax>();
        foreach (var identifier in identifiers)
        {
            var symbolInfo = model.GetSymbolInfo(identifier);
            var symbol = symbolInfo.Symbol;

            if (symbol != null)
            {
                // 记录使用（通过类名.成员名匹配）
                _used.Add($"{symbol.ContainingType?.Name}.{symbol.Name}");
                _used.Add(symbol.Name); // 也记录简单名称
            }
        }

        // 2. 查找方法调用（包括 lambda 表达式中的）
        var invocations = root.DescendantNodes().OfType<InvocationExpressionSyntax>();
        foreach (var invocation in invocations)
        {
            var symbolInfo = model.GetSymbolInfo(invocation);
            var symbol = symbolInfo.Symbol;

            if (symbol != null)
            {
                _used.Add($"{symbol.ContainingType?.Name}.{symbol.Name}");
                _used.Add(symbol.Name);
            }
        }

        // 3. 查找成员访问表达式
        var memberAccesses = root.DescendantNodes().OfType<MemberAccessExpressionSyntax>();
        foreach (var memberAccess in memberAccesses)
        {
            var symbolInfo = model.GetSymbolInfo(memberAccess);
            var symbol = symbolInfo.Symbol;

            if (symbol != null)
            {
                _used.Add($"{symbol.ContainingType?.Name}.{symbol.Name}");
                _used.Add(symbol.Name);
            }
        }
    }

    private bool ShouldCheck(ISymbol symbol)
    {
        // 如果成员有 Attribute 标注，跳过（可能被框架/序列化器使用）
        if (symbol.GetAttributes().Length > 0)
            return false;

        // 跳过接口中的成员定义
        var containingType = symbol.ContainingType;
        if (containingType != null && containingType.TypeKind == TypeKind.Interface)
            return false;

        // 跳过实现接口的方法（这些方法由框架调用）
        if (symbol is IMethodSymbol method)
        {
            // 跳过 override 和 virtual 方法
            if (method.IsOverride || method.IsVirtual)
                return false;

            // 检查方法是否实现了接口
            var interfaceImplementations = containingType?.AllInterfaces;
            if (interfaceImplementations != null && containingType != null)
            {
                foreach (var iface in interfaceImplementations)
                {
                    var interfaceMembers = iface.GetMembers().OfType<IMethodSymbol>();
                    foreach (var ifaceMethod in interfaceMembers)
                    {
                        var implementation = containingType.FindImplementationForInterfaceMember(ifaceMethod);
                        if (SymbolEqualityComparer.Default.Equals(implementation, symbol))
                        {
                            // 这是接口实现方法，跳过检测
                            return false;
                        }
                    }
                }
            }
        }

        // 跳过 Startup 类的特殊方法
        if (containingType != null && containingType.Name == "Startup")
        {
            var name = symbol.Name;
            if (name == "ConfigureServices" || name == "Configure" || name == "ConfigureContainer")
                return false;
        }

        // 跳过 Controller、中间件、过滤器中的所有成员
        if (containingType != null)
        {
            var typeName = containingType.Name;
            var baseTypes = GetBaseTypes(containingType);

            // Controller 相关
            if (typeName.EndsWith("Controller") ||
                baseTypes.Any(t => t.EndsWith("Controller") || t.EndsWith("ControllerBase")))
                return false;

            // 中间件相关
            if (typeName.EndsWith("Middleware") ||
                baseTypes.Any(t => t.EndsWith("Middleware")))
                return false;

            // 过滤器相关
            if (typeName.EndsWith("Filter") ||
                typeName.EndsWith("Attribute") ||
                baseTypes.Any(t => t.Contains("Filter") || t.Contains("Attribute")))
                return false;
        }

        // 跳过特殊方法
        var name2 = symbol.Name;
        if (name2 == "Main" || name2.StartsWith("On") || name2.EndsWith("Changed"))
            return false;

        return true;
    }

    private List<string> GetBaseTypes(INamedTypeSymbol type)
    {
        var result = new List<string>();
        var current = type.BaseType;

        while (current != null)
        {
            result.Add(current.Name);
            current = current.BaseType;
        }

        return result;
    }
}

// 数据模型
public class AnalysisResult
{
    public int TotalDefined { get; set; }
    public int TotalUsed { get; set; }
    public List<DeadCodeItem> DeadCodeItems { get; set; } = new();
}

public class DeadCodeItem
{
    public string Symbol { get; set; } = "";
}
