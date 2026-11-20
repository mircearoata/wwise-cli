package product

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/mircearoata/wwise-cli/lib/wwise/client"
	"github.com/mircearoata/wwise-cli/utils"
	"github.com/spf13/viper"
)

type WwiseProduct struct {
	Client      *client.WwiseClient
	ProductName string
}

func NewWwiseProduct(client *client.WwiseClient, productName string) *WwiseProduct {
	return &WwiseProduct{
		Client:      client,
		ProductName: productName,
	}
}

func (p *WwiseProduct) GetInfo() (ProductInfo, error) {
	payload, err := p.Client.SendRequest("GET", "/products/versions/?category="+p.ProductName, nil)
	if err != nil {
		return ProductInfo{}, fmt.Errorf("failed to get product info: %w", err)
	}

	var data struct {
		Data ProductInfo `json:"data"`
	}
	err = json.Unmarshal([]byte(payload), &data)
	if err != nil {
		return ProductInfo{}, fmt.Errorf("failed to unmarshal product info: %w", err)
	}

	return data.Data, nil
}

func (p *WwiseProduct) GetVersion(version string) (*WwiseProductVersion, error) {
	version = strings.TrimPrefix(version, p.ProductName+".")
	cacheDir := filepath.Join(viper.GetString("cache-dir"), p.ProductName, version)
	pv := &WwiseProductVersion{
		Product:   p,
		VersionId: version,
		Dir:       cacheDir,
		downloadedInfo: &WwiseVersionDownloadedInfo{
			Files:  []string{},
			Groups: []Group{},
		},
	}
	err := pv.readDownloadedInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to read downloaded info: %w", err)
	}

	return pv, nil
}

type WwiseProductVersion struct {
	Product        *WwiseProduct
	VersionId      string
	Dir            string
	downloadedInfo *WwiseVersionDownloadedInfo
}

func (v *WwiseProductVersion) Save() error {
	file := filepath.Join(v.Dir, "info.json")

	infoJson, err := json.MarshalIndent(v.downloadedInfo, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal downloaded info: %w", err)
	}

	if err := os.WriteFile(file, infoJson, 0755); err != nil {
		return fmt.Errorf("failed to write downloaded info: %w", err)
	}
	return nil
}

func (v *WwiseProductVersion) readDownloadedInfo() error {
	file := filepath.Join(v.Dir, "info.json")

	_, err := os.Stat(file)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to stat downloaded info file: %w", err)
		}

		_, err := os.Stat(v.Dir)
		if err != nil {
			if !os.IsNotExist(err) {
				return fmt.Errorf("failed to read cache directory: %w", err)
			}

			err = os.MkdirAll(v.Dir, 0755)
			if err != nil {
				return fmt.Errorf("failed to create cache directory: %w", err)
			}
		}

		if err := v.Save(); err != nil {
			return fmt.Errorf("failed to save empty downloaded info: %w", err)
		}

		return nil
	}

	downloadedInfoData, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("failed to read downloaded info: %w", err)
	}

	if err := json.Unmarshal(downloadedInfoData, v.downloadedInfo); err != nil {
		return fmt.Errorf("failed to unmarshal downloaded info: %w", err)
	}

	return nil
}

func (v *WwiseProductVersion) GetInfo() (ProductVersionInfo, error) {
	payload, err := v.Product.Client.SendRequest("GET", "/products/versions/"+v.Product.ProductName+"."+strings.ReplaceAll(v.VersionId, ".", "_"), nil)
	if err != nil {
		return ProductVersionInfo{}, fmt.Errorf("failed to get product version info: %w", err)
	}

	var data struct {
		Data ProductVersionInfo `json:"data"`
	}
	err = json.Unmarshal([]byte(payload), &data)
	if err != nil {
		return ProductVersionInfo{}, fmt.Errorf("failed to unmarshal product version info: %w", err)
	}

	return data.Data, nil
}

func (v *WwiseProductVersion) IsFileDownloaded(file string) bool {
	return v.downloadedInfo.IsFileDownloaded(file)
}

func (v *WwiseProductVersion) DownloadOrCache(file File, progressWriter io.Writer) error {
	if v.downloadedInfo.IsFileDownloaded(file.Name) {
		return nil
	}

	fileResp, err := http.Get(file.URL)
	if err != nil {
		return fmt.Errorf("failed to download integration file: %w", err)
	}
	defer fileResp.Body.Close()

	if fileResp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: %s returned status %d (%s)", file.Name, fileResp.StatusCode, fileResp.Status)
	}

	var reader io.Reader = fileResp.Body
	if progressWriter != nil {
		reader = io.TeeReader(reader, progressWriter)
	}
	reader = bufio.NewReaderSize(reader, 5*1024*1024)

	err = utils.ExtractTarXz(reader, v.Dir)
	if err != nil {
		return fmt.Errorf("failed to extract integration file: %w", err)
	}

	v.downloadedInfo.Files = append(v.downloadedInfo.Files, file.Name)
	for _, group := range file.Groups {
		if v.downloadedInfo.IsGroupDownloaded(group.GroupID, group.GroupValueID) {
			continue
		}
		v.downloadedInfo.Groups = append(v.downloadedInfo.Groups, group)
	}
	if err := v.Save(); err != nil {
		return fmt.Errorf("failed to save downloaded info: %w", err)
	}

	return nil
}

type WwiseVersionDownloadedInfo struct {
	Files  []string `json:"files"`
	Groups []Group  `json:"groups"`
}

func (info *WwiseVersionDownloadedInfo) IsFileDownloaded(file string) bool {
	for _, downloadedFile := range info.Files {
		if downloadedFile == file {
			return true
		}
	}
	return false
}

func (info *WwiseVersionDownloadedInfo) IsGroupDownloaded(groupId string, groupValue string) bool {
	for _, downloadedGroup := range info.Groups {
		if downloadedGroup.GroupID == groupId && downloadedGroup.GroupValueID == groupValue {
			return true
		}
	}
	return false
}
