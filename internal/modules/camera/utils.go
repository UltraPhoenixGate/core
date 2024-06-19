package camera

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"

	"github.com/sirupsen/logrus"
	"github.com/use-go/onvif"
)

func init() {
	// 创建snapshots目录
	if _, err := os.Stat("snapshots"); os.IsNotExist(err) {
		os.Mkdir("snapshots", os.ModePerm)
	}
}

func captureSnapshot(streamURL string) (string, error) {
	md5 := md5.New()
	outputFile := fmt.Sprintf("snapshots/%s.jpg", hex.EncodeToString(md5.Sum([]byte(streamURL))))
	// remove file if exists
	if _, err := os.Stat(outputFile); err == nil {
		os.Remove(outputFile)
	}
	cmd := exec.Command("ffmpeg", "-i", streamURL, "-vf", "select='eq(n\\,0)'", "-frames:v", "1", outputFile)
	logrus.Infof("Running command: %v", cmd.String())
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

func getXMLNode(xmlBody string, nodeName string) (*xml.Decoder, *xml.StartElement, string) {

	xmlBytes := bytes.NewBufferString(xmlBody)
	decodedXML := xml.NewDecoder(xmlBytes)

	for {
		token, err := decodedXML.Token()
		if err != nil {
			break
		}
		switch et := token.(type) {
		case xml.StartElement:
			if et.Name.Local == nodeName {
				return decodedXML, &et, ""
			}
		}
	}
	return nil, nil, "error in NodeName"
}

func CallDeviceMethod(dev *onvif.Device, request interface{}, responseTag string, response interface{}) error {
	services := dev.GetServices()
	logrus.WithField("device", services).Info("Found onvif device")
	resp, err := dev.CallMethod(request)
	if err != nil {
		return fmt.Errorf("failed to call method: %w", err)
	}
	defer resp.Body.Close()

	bs, _ := io.ReadAll(resp.Body)
	xmlStr, et, errFunc := getXMLNode(string(bs), responseTag)
	if errFunc != "" {
		return fmt.Errorf("failed to get response: %v", errFunc)
	}
	if err := xmlStr.DecodeElement(response, et); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}
	return nil
}

func GetDeviceXAddr(dev *onvif.Device) string {
	services := dev.GetServices()
	for _, service := range services {
		if url, err := url.Parse(service); err == nil {
			return url.Host
		}
	}
	return ""
}
