package inputmethod

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

type WordEntry struct {
	Word   string `json:"text"`
	Pinyin string `json:"pinyin"`
	Weight int    `json:"weight"`
}

const weight = 9000

var term_resource = os.Getenv("term_resource")

func decodeLineToUTF8(rawLine []byte) string {
	rawLine = bytes.TrimPrefix(rawLine, []byte{0xEF, 0xBB, 0xBF})
	if utf8.Valid(rawLine) {
		return string(rawLine)
	}

	decoded, _, err := transform.Bytes(simplifiedchinese.GB18030.NewDecoder(), rawLine)
	if err == nil && utf8.Valid(decoded) {
		return string(decoded)
	}

	return string(rawLine)
}

// rawLine 表示词库文件中的一行。
// isData 为 true 时代表 "词\t拼音\t权重" 格式的词条行。
type rawLine struct {
	text   string
	isData bool
	word   string
}

func formatEntry(entry WordEntry) string {
	return fmt.Sprintf("%s\t%s\t%d", entry.Word, entry.Pinyin, entry.Weight)
}

func SetTerm(term *WordEntry) error {
	if term == nil {
		return fmt.Errorf("term 不能为空")
	}

	if term_resource == "" {
		return fmt.Errorf("环境变量 term_resource 未设置")
	}

	if term.Word == "" || term.Pinyin == "" {
		return fmt.Errorf("term 的 word 或 pinyin 不能为空")
	}

	if term.Weight <= 0 {
		term.Weight = weight
	}

	file, err := os.OpenFile(term_resource, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("词库路劲设置有问题,请检查环境变量 term_resource")
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var lines []rawLine

	for scanner.Scan() {
		line := decodeLineToUTF8(append([]byte(nil), scanner.Bytes()...))
		line = strings.TrimSuffix(line, "\r")

		parts := strings.Split(line, "\t")
		// 两列及以上且首列不是注释的，视为词条行（词\t拼音\t权重 / 词\t拼音）
		if len(parts) >= 2 && parts[0] != "" && !strings.HasPrefix(parts[0], "#") {
			lines = append(lines, rawLine{text: line, isData: true, word: parts[0]})
		} else {
			// 注释、表头（如 #@/db_name）、空行等，原样保留
			lines = append(lines, rawLine{text: line})
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("读取词库文件失败: %w", err)
	}

	// 替换同名词条；文件里没有则追加到末尾
	added := false
	var out []string
	for _, l := range lines {
		if l.isData && l.word == term.Word {
			if !added {
				out = append(out, formatEntry(*term))
				added = true
			}
			continue
		}
		out = append(out, l.text)
	}
	if !added {
		out = append(out, formatEntry(*term))
	}

	err = file.Truncate(0)
	if err != nil {
		return fmt.Errorf("清空文件失败: %w", err)
	}
	if _, err := file.Seek(0, 0); err != nil {
		return fmt.Errorf("重置文件指针失败: %w", err)
	}

	writer := bufio.NewWriter(file)
	for _, line := range out {
		if _, err := writer.WriteString(line + "\n"); err != nil {
			return fmt.Errorf("写入文件失败: %w", err)
		}
	}

	if err := writer.Flush(); err != nil {
		return fmt.Errorf("刷新缓冲区失败: %w", err)
	}
	return nil
}
