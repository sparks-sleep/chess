package res

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

//FileToByte 把文件转成Byte数组
func FileToByte(inPath, outPath string) error {
	dir, err := ioutil.ReadDir(inPath)
	if err != nil {
		return err
	}

	fOut, err := os.Create(outPath + "/resources.go")
	if err != nil {
		return err
	}
	defer fOut.Close()

	//写入包名
	if _, err := fmt.Fprintf(fOut, "package chess\n\n"); err != nil {
		return err
	}

	//初始化map
	if _, err := fmt.Fprintf(fOut, "var resMap = map[int][]byte{\n"); err != nil {
		return err
	}
	//目录下得所有文件
	for _, fIn := range dir {
		//生成变量名
		varName := ""
		//HasSuffix判断字符串fIn.Name()是否以.png结尾
		if ok := strings.HasSuffix(fIn.Name(), ".png"); ok {
			//TrimSuffix 返回没有提供的尾随后缀字符串的".png"。
			//如果 fIn.Name() 不以".png" suffix 结尾，则 fIn.Name() 原样返回。
			varName = "Img" + strings.TrimSuffix(fIn.Name(), ".png")
		} else if ok := strings.HasSuffix(fIn.Name(), ".wav"); ok {
			varName = "Music" + strings.TrimSuffix(fIn.Name(), ".wav")
		} else {
			continue
		}

		//打开输入文件
		f, err := os.Open(inPath + "/" + fIn.Name())
		if err != nil {
			return err
		}
		defer f.Close()
		bs, err := ioutil.ReadAll(f)
		if err != nil {
			return err
		}

		//写入输出文件
		if _, err := fmt.Fprintf(fOut, "\t%s: []byte(%q),\n", varName, bs); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintf(fOut, "}"); err != nil {
		return err
	}

	return nil
}
