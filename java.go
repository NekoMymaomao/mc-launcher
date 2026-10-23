package main

// java.go — Java 探测与安装
import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

type JavaInfo struct {
	Path    string
	Major   int
	Version string
}

// javaVersionRe 匹配 `java -version` 的版本号：
//   openjdk version "17.0.2" 2022-01-18
//   java version "1.8.0_392"
//   openjdk 21.0.1 2023-10-17
// 锚定 "version" 关键字或 "openjdk X.Y" 行首，避免把日期里的
// "2022" 之类数字当成主版本号。
var javaVersionRe = regexp.MustCompile(`(?i)\b(?:version|openjdk)\s+(?:["']?)(\d+)(?:\.(\d+))?`)

func scanJava() []JavaInfo {
	var out []JavaInfo
	seen := map[string]bool{}
	add := func(p string) {
		p = strings.TrimSpace(p)
		if p == "" || seen[p] {
			return
		}
		seen[p] = true
		ji := probeJava(p)
		if ji.Major > 0 {
			out = append(out, ji)
		}
	}
	if cfg != nil && cfg.JavaPath != "" {
		add(cfg.JavaPath)
	}
	if v := os.Getenv("JAVA_HOME"); v != "" {
		add(filepath.Join(v, "bin", "java"))
		add(filepath.Join(v, "bin", "java.exe"))
	}
	javaDir := filepath.Join(baseDir, "java")
	if info, err := os.Stat(javaDir); err == nil && info.IsDir() {
		walkJava(javaDir, add)
	}
	// PATH 查找：Windows 没有 which
	if runtimeGOOS == "windows" {
		if out, err := exec.Command("where", "java").Output(); err == nil {
			for _, line := range strings.Split(string(out), "\n") {
				add(line)
			}
		}
	} else {
		if out, err := exec.Command("which", "java").Output(); err == nil {
			for _, line := range strings.Split(string(out), "\n") {
				add(line)
			}
		}
	}
	for _, root := range commonJavaRoots() {
		walkJava(root, add)
	}
	return out
}

// walkJava 安全遍历目录寻找 java 可执行文件。
// filepath.Walk 会跟随目录软链接，若系统里存在循环链接
// （A -> B -> A）会无限递归、吃满 CPU；因此：
//  1. 目录若是软链接则 SkipDir（不进入链接目标）；
//  2. 用 lstat 路径集合做环检测，重复出现的目录直接跳过；
//  3. 深度限制 12 层，再深的目录不可能正常放着 Java。
func walkJava(root string, add func(string)) {
	seen := map[string]bool{}
	walkDirLimited(root, 12, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			if isDirSymlink(p) {
				return filepath.SkipDir
			}
			real, err := filepath.EvalSymlinks(p)
			if err == nil {
				if seen[real] {
					log.Printf("java: 跳过循环目录 %s -> %s", p, real)
					return filepath.SkipDir
				}
				seen[real] = true
			}
			return nil
		}
		b := filepath.Base(p)
		if b == "java" || b == "java.exe" {
			add(p)
		}
		return nil
	})
}

func isDirSymlink(p string) bool {
	fi, err := os.Lstat(p)
	return err == nil && fi.Mode()&os.ModeSymlink != 0
}

// walkDirLimited 带深度上限的目录遍历（超过 maxDepth 层的子目录不再进入）
func walkDirLimited(root string, maxDepth int, fn filepath.WalkFunc) {
	filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
		if err != nil || info == nil {
			return nil
		}
		if info.IsDir() && p != root {
			depth := strings.Count(strings.TrimPrefix(p, root), string(os.PathSeparator))
			if depth > maxDepth {
				return filepath.SkipDir
			}
		}
		return fn(p, info, nil)
	})
}

func commonJavaRoots() []string {
	switch runtimeGOOS {
	case "windows":
		return []string{
			`C:\Program Files\Java`,
			`C:\Program Files\Eclipse Adoptium`,
			`C:\Program Files\Microsoft\OpenJDK`,
			`C:\Program Files\Amazon Corretto`,
		}
	default:
		return []string{"/usr/lib/jvm", "/usr/java", "/opt/java", "/opt/jdk"}
	}
}

// probeJava 执行 java -version 并解析主版本号
func probeJava(path string) JavaInfo {
	ji := JavaInfo{Path: path}
	if info, err := os.Stat(path); err != nil || info.IsDir() {
		return ji
	}
	out, err := exec.Command(path, "-version").CombinedOutput()
	if err != nil {
		log.Printf("java: 探测 %s 失败: %v", path, err)
		return ji
	}
	text := string(out)
	ji.Version = strings.TrimSpace(text)
	ji.Major = parseJavaMajor(text)
	if ji.Major == 0 {
		log.Printf("java: 无法从 %s 的输出解析主版本: %q", path, truncateForLog(ji.Version))
	}
	return ji
}

func truncateForLog(s string) string {
	s = strings.TrimSpace(s)
	if len(s) > 120 {
		return s[:120] + "..."
	}
	return s
}

// parseJavaMajor 解析 java -version 输出里的主版本号
func parseJavaMajor(out string) int {
	for _, line := range strings.Split(out, "\n") {
		if !strings.Contains(line, "version") && !strings.Contains(line, "openjdk") {
			continue
		}
		m := javaVersionRe.FindStringSubmatch(line)
		if len(m) < 2 {
			continue
		}
		major := parseIntClamp(m[1])
		if major == 1 && len(m) > 2 && m[2] != "" {
			// 1.8.0_392 -> 8
			return parseIntClamp(m[2])
		}
		if major > 0 {
			return major
		}
	}
	return 0
}

// pickJava 选择满足最低主版本的 Java；没有则返回 nil
func pickJava(needMajor int) *JavaInfo {
	all := scanJava()
	if len(all) == 0 {
		return nil
	}
	var best *JavaInfo
	for i := range all {
		j := &all[i]
		if j.Major < needMajor {
			continue
		}
		// 优先取满足要求里最小的（别拿 Java 21 去开 1.8）
		if best == nil || j.Major < best.Major {
			best = j
		}
	}
	return best
}

func parseIntClamp(s string) int {
	var n int
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		} else {
			break
		}
	}
	return n
}

func installJava(major int) (string, error) {
	if major <= 0 {
		major = 17
	}
	ext := "tar.gz"
	if runtimeGOOS == "windows" {
		ext = "zip"
	}
	archive := "temurin-" + itoa(major) + "." + ext
	url := "https://api.adoptium.net/v3/binary/latest/" + itoa(major) + "/ga/" +
		runtimeGOOS + "/" + runtimeGOARCH + "/jre/hotspot/normal/eclipse"
	dest := filepath.Join(baseDir, "java", archive)
	if err := downloadFile(mirrorUrl(url, cfg.DownloadSource), dest, "", nil); err != nil {
		return "", err
	}
	target := filepath.Join(baseDir, "java", "temurin-"+itoa(major))
	if ext == "zip" {
		if err := unzip(dest, target, nil); err != nil {
			return "", err
		}
	} else {
		if err := untarGz(dest, target); err != nil {
			return "", err
		}
	}
	os.Remove(dest)
	if found := findJavaBinary(target); found != "" {
		cfg.JavaPath = found
		cfg.save()
		return found, nil
	}
	return "", fmt.Errorf("解压后未找到 java 可执行文件")
}

func findJavaBinary(dir string) string {
	cands := []string{"bin/java", "bin/java.exe", "jre/bin/java", "jre/bin/java.exe"}
	for _, c := range cands {
		p := filepath.Join(dir, c)
		if fileExists(p) {
			return p
		}
	}
	var found string
	// 一旦找到就把后续目录全部 SkipDir，避免无谓遍历；
	// 目录软链接同样不进入，防止循环。
	filepath.Walk(dir, func(p string, info os.FileInfo, err error) error {
		if err != nil || info == nil {
			return nil
		}
		if found != "" {
			// SkipDir 只对目录生效，文件时返回 nil 即可（回调很快随父目录结束）
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if info.IsDir() {
			if isDirSymlink(p) {
				return filepath.SkipDir
			}
			depth := strings.Count(strings.TrimPrefix(p, dir), string(os.PathSeparator))
			if depth > 8 {
				return filepath.SkipDir
			}
			return nil
		}
		b := filepath.Base(p)
		if b == "java" || b == "java.exe" {
			found = p
		}
		return nil
	})
	return found
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var out []byte
	for n > 0 {
		out = append([]byte{byte('0' + n%10)}, out...)
		n /= 10
	}
	if neg {
		return "-" + string(out)
	}
	return string(out)
}
