package dotnets

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func CreateDotNetProject(projectName string) error {
	if projectName == "" {
		return nil
	}
	fmt.Printf("正在创建 .NET 项目 (%s)\n", projectName)

	// 定义层级名称
	domainLayer := projectName + ".Domain"
	appLayer := projectName + ".Application"
	infraLayer := projectName + ".Infrastructure"
	apiLayer := projectName + ".Api"
	sharedLayer := projectName + ".Shared"

	// 定义项目路径
	domainPath := "src/" + domainLayer
	appPath := "src/" + appLayer
	infraPath := "src/" + infraLayer
	apiPath := "src/" + apiLayer
	sharedPath := "src/" + sharedLayer

	// 定义 .csproj 文件路径
	domain := fmt.Sprintf("%s/%s.csproj", domainPath, domainLayer)
	app := fmt.Sprintf("%s/%s.csproj", appPath, appLayer)
	infra := fmt.Sprintf("%s/%s.csproj", infraPath, infraLayer)
	api := fmt.Sprintf("%s/%s.csproj", apiPath, apiLayer)
	shared := fmt.Sprintf("%s/%s.csproj", sharedPath, sharedLayer)

	commands := [][]string{
		// 创建解决方案
		{"dotnet", "new", "sln", "-n", projectName},

		// 创建各层项目
		{"dotnet", "new", "classlib", "-n", domainLayer, "-o", domainPath},
		{"dotnet", "new", "classlib", "-n", appLayer, "-o", appPath},
		{"dotnet", "new", "classlib", "-n", infraLayer, "-o", infraPath},
		{"dotnet", "new", "webapi", "-n", apiLayer, "-o", apiPath},
		{"dotnet", "new", "classlib", "-n", sharedLayer, "-o", sharedPath},

		// 添加到解决方案
		{"dotnet", "sln", "add", domain},
		{"dotnet", "sln", "add", app},
		{"dotnet", "sln", "add", infra},
		{"dotnet", "sln", "add", api},
		{"dotnet", "sln", "add", shared},

		// 添加项目引用
		{"dotnet", "add", app, "reference", domain},
		{"dotnet", "add", app, "reference", shared},
		{"dotnet", "add", domain, "reference", shared},
		{"dotnet", "add", infra, "reference", domain},
		{"dotnet", "add", infra, "reference", shared},
		{"dotnet", "add", api, "reference", app},
		{"dotnet", "add", api, "reference", infra},

		//添加基本的nuget包
		{"dotnet", "add", api, "package", "NLog.Web.AspNetCore"},
		{"dotnet", "add", api, "package", "Microsoft.Extensions.Logging.Console"},
		{"dotnet", "add", api, "package", "Microsoft.AspNetCore.OpenApi"},
		{"dotnet", "add", infra, "package", "Microsoft.Extensions.Configuration"},
		{"dotnet", "add", infra, "package", "Dapper"},
		{"dotnet", "add", domain, "package", "AutoMapper"},
		{"dotnet", "add", app, "package", "MediatR"},
	}

	// 逐个执行
	for _, cmdArgs := range commands {
		fmt.Printf("执行: %v\n", cmdArgs)

		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("命令执行失败 %v: %v", cmdArgs, err)
		}
	}
	// 生成必要的代码文件
	err := createProgramCsFile(apiPath, appLayer)
	if err != nil {
		return fmt.Errorf("创建 Api/Program.cs 失败: %v", err)
	}

	err = createStartupCsFile(apiPath, appLayer)
	if err != nil {
		return fmt.Errorf("创建 Api/Startup.cs 失败: %v", err)
	}

	err = createInfrastructureServiceCollectionCsFile(infraPath, infraLayer)
	if err != nil {
		return fmt.Errorf("创建 Infrastructure/ServiceCollectionExtensions.cs 失败: %v", err)
	}

	err = createConfigFiles(apiPath)
	if err != nil {
		return fmt.Errorf("创建配置文件失败: %v", err)
	}
	fmt.Printf("✓ .NET 项目 %s 创建完成\n", projectName)
	return nil
}

func createProgramCsFile(apiPath string, appLayer string) error {
	apiProgramContent :=
		`using Microsoft.AspNetCore.Builder;
		using NLog.Extensions.Logging;
		using Microsoft.Extensions.DependencyInjection;

		namespace ` + appLayer + ` ;
		
		public class Program
		{
			public static void Main(string[] args)
			{
				 CreateHostBuilder(args).Build().Run();
			}

			public static IHostBuilder CreateHostBuilder(string[] args) =>
			
				Host.CreateDefaultBuilder(args)
				.ConfigureWebHostDefaults(webBuilder =>
				{
					webBuilder.ConfigureKestrel(kestrel =>
					{
						kestrel.ListenAnyIP(80);
					});
					webBuilder.UseStartup<Startup>();
				})
				.ConfigureAppConfiguration((ctx, config) =>{
					config.SetBasePath(Directory.GetCurrentDirectory())
            		.AddJsonFile("Configs/appsettings.json", optional: false, reloadOnChange: true);
				})
				.ConfigureLogging((context, logging) =>
				{
					logging.AddNLog("Configs/nlog.config");
					logging.AddSimpleConsole();
				});
		}`

	programFilePath := filepath.Join(apiPath, "Program.cs")
	err := os.WriteFile(programFilePath, []byte(apiProgramContent), 0644)
	if err != nil {
		return fmt.Errorf("生成 Api/Program.cs 失败: %v", err)
	}
	return nil
}

func createStartupCsFile(apiPath string, appLayer string) error {
	apiStartupContent := `
	` + ` 
		namespace ` + appLayer + `;

		public class Startup(IConfiguration configuration, IWebHostEnvironment hostingEnvironment)
		{
		  	public IConfiguration Configuration { get; } = configuration;
   			public IWebHostEnvironment HostingEnvironment { get; } = hostingEnvironment;

			public void Configure(IApplicationBuilder app, IWebHostEnvironment env)
			{
				app.UseRouting();
				app.UseEndpoints(endpoints =>
				{
					endpoints.MapControllers();
				});
			}
		}
			
	`
	startUpFilePath := filepath.Join(apiPath, "Startup.cs")
	err := os.WriteFile(startUpFilePath, []byte(apiStartupContent), 0644)
	if err != nil {
		return fmt.Errorf("生成 Api/Startup.cs 失败: %v", err)
	}
	return nil
}

func createInfrastructureServiceCollectionCsFile(infraPath string, infraLayer string) error {
	infraServiceCollectionContent := `
		using Microsoft.Extensions.Configuration;
		using Microsoft.Extensions.DependencyInjection;
	` + `
		namespace ` + infraLayer + `;

		public static class ServiceCollectionExtensions
		{
			public static IServiceCollection AddInfrastructure(this IServiceCollection services, IConfiguration configuration)
			{
				 return services;
			}
		}
	`
	serviceCollectionFilePath := filepath.Join(infraPath, "ServiceCollectionExtensions.cs")
	err := os.WriteFile(serviceCollectionFilePath, []byte(infraServiceCollectionContent), 0644)
	if err != nil {
		return fmt.Errorf("生成 Infrastructure/ServiceCollectionExtensions.cs 失败: %v", err)
	}
	return nil
}

func createConfigFiles(apiPath string) error {
	configsDir := filepath.Join(apiPath, "Configs")
	err := os.MkdirAll(configsDir, 0755)
	if err != nil {
		return fmt.Errorf("创建 Configs 目录失败: %v", err)
	}

	oldAppSettings := filepath.Join(apiPath, "appsettings.json")
	newAppSettings := filepath.Join(configsDir, "appsettings.json")
	if _, err := os.Stat(oldAppSettings); err == nil {
		os.Rename(oldAppSettings, newAppSettings)
	}

	oldAppSettingsDev := filepath.Join(apiPath, "appsettings.Development.json")
	newAppSettingsDev := filepath.Join(configsDir, "appsettings.Development.json")
	if _, err := os.Stat(oldAppSettingsDev); err == nil {
		os.Rename(oldAppSettingsDev, newAppSettingsDev)
	}

	nlogConfigPath := filepath.Join(configsDir, "nlog.config")

	nlogContent := `
		<?xml version="1.0" ?>
<nlog xmlns="http://www.nlog-project.org/schemas/NLog.xsd"
      xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
      autoReload="true">

  <targets>
      <target name="normalLogFile" xsi:type="File" fileName="logs/normal-${shortdate}.log"  encoding="UTF-8">
        <layout type="JsonLayout" includeAllProperties="true" >
          <attribute name="time" layout="${date:yyyy\-MM\-dd HH\:mm\:ss.fff}" />
          <attribute name="level" layout="${level}" />
          <attribute name ="logger" layout="${logger}"/>
          <attribute name="msg" layout="${message}" />
          <attribute name="exception" layout="${exception:format=tostring}" />
          <attribute name="traceid" layout="${aspnet-traceidentifier}" />
        </layout>
      </target>
      <target  name="accessLogFile" xsi:type="File" fileName="logs/access-${shortdate}.log" layout="${message} ${exception:format=tostring}" >
      </target>
      <target name="callLogFile" xsi:type="File" fileName="logs/call-${shortdate}.log" layout="${message} ${exception:format=tostring}" >
      </target>
    <target xsi:type="Null" name="blackhole" />

    <!--输出至Debugger-->
    <target name="debugger" xsi:type="Debugger" layout="NLog: ${date:format=HH\:mm\:ss} | ${message}" />
    <!--输出至控制台-->
    <target name=" " xsi:type="ColoredConsole" layout="${longdate} ${level:uppercase=true} ${logger} ${threadid} ${message} ${exception:format=tostring}"></target>
  </targets>

  <rules>
    <!--访问日志-->
    <logger name="accessLog" minlevel="Debug" writeTo="accessLogFile" final="true" />
    <!--调用日志-->
    <logger name="callLog" minlevel="Debug" writeTo="callLogFile" final="true" />
    <!--这里记录EF的SQL日志-->
    <!--<logger name="Microsoft.EntityFrameworkCore.Database.Command" minlevel="Info" writeTo="normalLogFile" final="true"  />-->
    <!--屏蔽微软框架的日志-->
    <logger name="Microsoft.*" maxlevel="Info" writeTo="blackhole" final="true" />
    <logger name="RawRabbit.*" maxlevel="Info" writeTo="blackhole" final="true" />
    <logger name="*" minlevel="Error" writeTo="console" />
    <!--普通日志-->
    <logger name="*" minlevel="Info" writeTo="normalLogFile" final="true"/>

  </rules>
</nlog>

	`
	err = os.WriteFile(nlogConfigPath, []byte(nlogContent), 0644)
	if err != nil {
		return fmt.Errorf("生成 nlog.config 失败: %v", err)
	}

	fmt.Println("✓ 已创建 Configs 目录并配置文件")
	return nil
}
