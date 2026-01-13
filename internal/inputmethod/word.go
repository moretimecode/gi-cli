package inputmethod

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strconv"
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

	exitsWord := make(map[string]WordEntry)

	file, err := os.OpenFile(term_resource, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("词库路劲设置有问题,请检查环境变量 term_resource")
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := decodeLineToUTF8(append([]byte(nil), scanner.Bytes()...))
		line = strings.TrimSuffix(line, "\r")

		if line == "" {
			continue
		}

		parts := strings.Split(line, "\t")
		if len(parts) < 3 {
			fmt.Printf("第 %d 行格式错误: %s\n", lineNum, line)
			continue
		}

		num, err := strconv.Atoi(parts[2])
		if err != nil {
			num = weight
		}

		entry := WordEntry{
			Word:   parts[0],
			Pinyin: parts[1],
			Weight: num,
		}

		exitsWord[entry.Word] = entry
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("读取词库文件失败: %w", err)
	}

	exitsWord[term.Word] = *term

	err = file.Truncate(0)
	if err != nil {
		return fmt.Errorf("清空文件失败: %w", err)
	}
	if _, err := file.Seek(0, 0); err != nil {
		return fmt.Errorf("重置文件指针失败: %w", err)
	}

	writer := bufio.NewWriter(file)

	for _, entry := range exitsWord {
		line := fmt.Sprintf("%s\t%s\t%d\n", entry.Word, entry.Pinyin, entry.Weight)
		_, err := writer.WriteString(line)
		if err != nil {
			return fmt.Errorf("写入文件失败: %w", err)
		}
	}

	if err := writer.Flush(); err != nil {
		return fmt.Errorf("刷新缓冲区失败: %w", err)
	}
	return nil
}
