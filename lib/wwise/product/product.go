package product

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/mircearoata/wwise-cli/lib/wwise/client"
	"github.com/mircearoata/wwise-cli/utils"
	"github.com/pkg/errors"
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
		return ProductInfo{}, errors.Wrap(err, "failed to get product info")
	}

	var data struct {
		Data ProductInfo `json:"data"`
	}
	err = json.Unmarshal([]byte(payload), &data)
	if err != nil {
		return ProductInfo{}, errors.Wrap(err, "failed to unmarshal product info")
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
		return nil, errors.Wrap(err, "failed to read downloaded info")
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
		return errors.Wrap(err, "failed to marshal downloaded info")
	}

	if err := os.WriteFile(file, infoJson, 0755); err != nil {
		return errors.Wrap(err, "failed to write downloaded info")
	}
	return nil
}

func (v *WwiseProductVersion) readDownloadedInfo() error {
	file := filepath.Join(v.Dir, "info.json")

	_, err := os.Stat(file)
	if err != nil {
		if !os.IsNotExist(err) {
			return errors.Wrap(err, "failed to stat downloaded info file")
		}

		_, err := os.Stat(v.Dir)
		if err != nil {
			if !os.IsNotExist(err) {
				return errors.Wrap(err, "failed to read cache directory")
			}

			err = os.MkdirAll(v.Dir, 0755)
			if err != nil {
				return errors.Wrap(err, "failed to create cache directory")
			}
		}

		if err := v.Save(); err != nil {
			return errors.Wrap(err, "failed to save empty downloaded info")
		}

		return nil
	}

	downloadedInfoData, err := os.ReadFile(file)
	if err != nil {
		return errors.Wrap(err, "failed to read downloaded info")
	}

	if err := json.Unmarshal(downloadedInfoData, v.downloadedInfo); err != nil {
		return errors.Wrap(err, "failed to unmarshal downloaded info")
	}

	return nil
}

func (v *WwiseProductVersion) GetInfo() (ProductVersionInfo, error) {
	payload, err := v.Product.Client.SendRequest("GET", "/products/versions/"+v.Product.ProductName+"."+strings.ReplaceAll(v.VersionId, ".", "_"), nil)
	if err != nil {
		return ProductVersionInfo{}, errors.Wrap(err, "failed to get product version info")
	}

	var data struct {
		Data ProductVersionInfo `json:"data"`
	}
	err = json.Unmarshal([]byte(payload), &data)
	if err != nil {
		return ProductVersionInfo{}, errors.Wrap(err, "failed to unmarshal product version info")
	}

	return data.Data, nil
}

func (v *WwiseProductVersion) DownloadOrCache(file File) error {
	if v.downloadedInfo.IsFileDownloaded(file.Name) {
		return nil
	}

	fileResp, err := http.Get(file.URL)
	if err != nil {
		return errors.Wrap(err, "failed to download integration file")
	}
	defer fileResp.Body.Close()

	err = utils.ExtractTarXz(fileResp.Body, v.Dir)
	if err != nil {
		return errors.Wrap(err, "failed to extract integration file")
	}

	v.downloadedInfo.Files = append(v.downloadedInfo.Files, file.Name)
	for _, group := range file.Groups {
		if v.downloadedInfo.IsGroupDownloaded(group.GroupID, group.GroupValueID) {
			continue
		}
		v.downloadedInfo.Groups = append(v.downloadedInfo.Groups, group)
	}
	v.Save()

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
