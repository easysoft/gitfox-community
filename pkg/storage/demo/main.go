package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/storage"
)

func main() {
	ctx := context.Background()

	// 使用临时目录作为存储根目录
	tmpDir, err := os.MkdirTemp("", "storage-demo")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmpDir)

	fmt.Println("=== Storage Package Demo ===")
	fmt.Printf("Storage root directory: %s\n\n", tmpDir)

	// 创建 filesystem 驱动
	driver, err := storage.NewDriver(ctx, "filesystem", map[string]interface{}{
		"rootdirectory": tmpDir,
	})
	if err != nil {
		panic(err)
	}

	// 创建 ContentStorage，设置路径前缀
	cs, err := storage.NewCommonContentStore(ctx, driver,
		storage.WithPrefix("/uploads"))
	if err != nil {
		panic(err)
	}

	fmt.Printf("ContentStorage prefix: %s\n\n", cs.PrefixDir())

	// 1. 写入文件
	fmt.Println("--- 1. Put: Write file ---")
	err = cs.Put(ctx, "/example.txt", []byte("Hello, World!"))
	if err != nil {
		panic(err)
	}
	fmt.Println("Written /example.txt with content: Hello, World!")

	// 2. 读取文件
	fmt.Println("\n--- 2. Get: Read file ---")
	content, err := cs.Get(ctx, "/example.txt")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Read /example.txt: %s\n", string(content))

	// 3. 获取文件信息
	fmt.Println("\n--- 3. Stat: Get file info ---")
	fileInfo, err := cs.Stat(ctx, "/example.txt")
	if err != nil {
		panic(err)
	}
	fmt.Printf("File: %s, Size: %d bytes, IsDir: %v, ModTime: %v\n",
		fileInfo.Path(), fileInfo.Size(), fileInfo.IsDir(), fileInfo.ModTime())

	// 4. 流式写入
	fmt.Println("\n--- 4. Save: Stream write ---")
	largeContent := []byte("This is a larger file content for stream writing demo.\nLine 2\nLine 3\n")
	reader := bytes.NewReader(largeContent)
	err = cs.Save(ctx, "/large-file.txt", reader, int64(len(largeContent)))
	if err != nil {
		panic(err)
	}
	fmt.Printf("Stream written /large-file.txt (%d bytes)\n", len(largeContent))

	// 5. 流式读取
	fmt.Println("\n--- 5. Open: Stream read ---")
	rc, err := cs.Open(ctx, "/large-file.txt")
	if err != nil {
		panic(err)
	}
	readContent, err := io.ReadAll(rc)
	rc.Close()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Stream read /large-file.txt:\n%s", string(readContent))

	// 6. 写入更多文件用于列表演示
	fmt.Println("\n--- 6. Create more files for listing ---")
	_ = cs.Put(ctx, "/docs/readme.md", []byte("# README"))
	_ = cs.Put(ctx, "/docs/guide.md", []byte("# Guide"))
	_ = cs.Put(ctx, "/images/logo.png", []byte("PNG DATA"))
	fmt.Println("Created: /docs/readme.md, /docs/guide.md, /images/logo.png")

	// 7. 列出 /docs 目录
	fmt.Println("\n--- 7. List: List /docs directory ---")
	docsFiles, err := cs.List(ctx, "/docs")
	if err != nil {
		panic(err)
	}
	fmt.Println("Files in /docs:")
	for _, f := range docsFiles {
		fmt.Printf("  - %s\n", f)
	}

	// 8. 列出 /images 目录
	fmt.Println("\n--- 8. List: List /images directory ---")
	imgFiles, err := cs.List(ctx, "/images")
	if err != nil {
		panic(err)
	}
	fmt.Println("Files in /images:")
	for _, f := range imgFiles {
		fmt.Printf("  - %s\n", f)
	}

	// 9. 删除文件
	fmt.Println("\n--- 9. Delete: Delete file ---")
	err = cs.Delete(ctx, "/example.txt")
	if err != nil {
		panic(err)
	}
	fmt.Println("Deleted /example.txt")

	// 验证删除
	_, err = cs.Get(ctx, "/example.txt")
	if err != nil {
		fmt.Printf("Verified: /example.txt no longer exists (error: %v)\n", err)
	}

	// 10. 使用 FileWriter 追加写入
	fmt.Println("\n--- 10. GetWriter: Append write ---")
	writer, err := cs.GetWriter(ctx, "/append-demo.txt", false)
	if err != nil {
		panic(err)
	}
	writer.Write([]byte("First line\n"))
	writer.Write([]byte("Second line\n"))
	writer.Commit(ctx)
	writer.Close()

	appendContent, _ := cs.Get(ctx, "/append-demo.txt")
	fmt.Printf("Content of /append-demo.txt:\n%s", string(appendContent))

	fmt.Println("\n=== Demo Complete ===")
}
