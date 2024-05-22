package camera

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"

	"github.com/sirupsen/logrus"
)

func captureSnapshot(streamURL string) (string, error) {
	md5 := md5.New()
	outputFile := fmt.Sprintf("snapshots/%s.jpg", md5.Sum([]byte(streamURL)))
	cmd := exec.Command("ffmpeg", "-i", streamURL, "-vf", "select='eq(n\\,0)'", "-frames:v", "1", outputFile)

	// 运行命令并获取输出
	output, err := cmd.CombinedOutput()
	if err != nil {
		logrus.Errorf("ffmpeg error: %v, output: %s", err, string(output))
		return "", fmt.Errorf("ffmpeg error: %v, output: %s", err, string(output))
	}

	return outputFile, nil
}

func genImageBase64(imagePath string) (string, error) {
	file, err := os.Open(imagePath)
	if err != nil {
		return "", err
	}
	bytes := make([]byte, 500000)
	n, err := file.Read(bytes)
	if err != nil {
		return "", err
	}
	output := base64.StdEncoding.EncodeToString(bytes[:n])
	output = fmt.Sprintf("data:image/jpeg;base64,%s", output)
	return output, nil
}
