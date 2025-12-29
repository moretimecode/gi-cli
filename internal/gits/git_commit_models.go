package gits

import "time"

// GitCommitMetrics 提交基础信息（核心表）
type GitCommitMetrics struct {
	CommitID       string // 提交ID（完整哈希）
	ShortHash      string // 短哈希（前7位，方便显示）
	Author         string // 作者名称
	Email          string // 作者邮箱（唯一标识作者）
	Timestamp      int64  // 提交时间戳（Unix时间戳，UTC）
	Message        string // 提交信息
	ParentCommitID string // 父提交ID（第一个父提交，用于追溯）
}

// CommitStatsMetrics 提交统计信息（汇总表）
type CommitStatsMetrics struct {
	CommitID      string // 提交ID
	FilesChanged  int    // 变更文件总数
	LinesAdded    int    // 总新增行数
	LinesDeleted  int    // 总删除行数
	LinesModified int    // 总修改行数（增+删）
	TotalLines    int    // 总修改行数（增+删）- 别名，用于查询
	FileAddCount  int    // 新增文件数
	FileDelCount  int    // 删除文件数
	IsMerge       bool   // 是否为合并提交
}

// AuthorMetrics 作者信息（维度表）
type AuthorMetrics struct {
	Email        string // 邮箱（主键）
	Name         string // 姓名（可能会变，记录最新的）
	FirstSeen    int64  // 首次提交时间戳
	LastSeen     int64  // 最后提交时间戳
	TotalCommits int    // 总提交数
	TotalAdded   int    // 总新增行数
	TotalDeleted int    // 总删除行数
}

// FileMetrics 文件变更信息（明细表）
type FileMetrics struct {
	CommitID      string // 提交ID
	FileName      string // 文件名称（不含路径）
	FilePath      string // 完整文件路径
	LinesModified int    // 修改的行数（增+删）
	TotalLines    int    // 当前提交后该文件的总行数
	ChangeType    string // 变更类型：added, modified, deleted, renamed
}

// BranchMetrics 分支信息
type BranchMetrics struct {
	BranchName   string // 分支名称
	IsDefault    bool   // 是否为默认分支（main/master）
	LastCommitID string // 最后一次提交ID
	IsMerged     bool   // 是否已合并到主分支
	MergedAt     int64  // 合并时间戳
	MergedInto   string // 合并到哪个分支
}

// BranchCommitMetrics 分支与提交的关系（多对多）
type BranchCommitMetrics struct {
	BranchName string // 分支名称
	CommitID   string // 提交ID
}

// ============================================
// 时间维度辅助结构（动态计算，不存储）
// ============================================

// CommitTime 提交时间维度（运行时根据本地时区动态计算）
type CommitTime struct {
	CommitID  string    // 提交ID
	LocalTime time.Time // 本地时间（根据系统时区）
	Year      int       // 年
	Month     int       // 月
	Day       int       // 日
	Hour      int       // 小时
	Weekday   int       // 星期几（0=Sunday, 1=Monday, ...）
	IsWeekend bool      // 是否周末
}

// ToLocalTime 将 GitCommitMetrics 转换为本地时间
func (g GitCommitMetrics) ToLocalTime() time.Time {
	return time.Unix(g.Timestamp, 0)
}

// ToLocalTime 将时间戳转换为本地时间维度
func ToLocalTime(commitID string, timestamp int64) CommitTime {
	t := time.Unix(timestamp, 0)
	weekday := int(t.Weekday())

	return CommitTime{
		CommitID:  commitID,
		LocalTime: t,
		Year:      t.Year(),
		Month:     int(t.Month()),
		Day:       t.Day(),
		Hour:      t.Hour(),
		Weekday:   weekday,
		IsWeekend: weekday == 0 || weekday == 6, // Sunday or Saturday
	}
}

///后续如果用户想定制化查询，或者数据量太大，那么就进行数据db 化编程sqlite,在隐藏文件夹下面，支持sqlite查询
