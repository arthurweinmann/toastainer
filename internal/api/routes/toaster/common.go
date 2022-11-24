package toaster

import (
	"context"
	"fmt"

	"github.com/microcosm-cc/bluemonday"
	"github.com/shurcooL/github_flavored_markdown"
	"github.com/toastate/toastainer/internal/api/common"
	"github.com/toastate/toastainer/internal/config"
	"github.com/toastate/toastainer/internal/db/objectdb"
	"github.com/toastate/toastainer/internal/db/objectdb/objectdberror"
	"github.com/toastate/toastainer/internal/db/redisdb"
	"github.com/toastate/toastainer/internal/model"
)

var xssPolicy = bluemonday.UGCPolicy()

func completeToasterDynFields(toaster *model.Toaster, renderMarkdown bool) *model.Toaster {
	if toaster.PictureExtension != "" {
		if config.APIPort != "" {
			toaster.PictureLink = "https://" + config.APIDomain + ":" + config.APIPort + "/toaster/picture/" + toaster.ID + "/picture" + toaster.PictureExtension
		} else {
			toaster.PictureLink = "https://" + config.APIDomain + "/toaster/picture/" + toaster.ID + "/picture" + toaster.PictureExtension
		}
	}

	if renderMarkdown && toaster.Readme != "" {
		toaster.Readme = string(xssPolicy.SanitizeBytes(github_flavored_markdown.Markdown([]byte(toaster.Readme))))
	}

	if toaster.CodeID != "" {
		if config.APIPort != "" {
			toaster.RunLink = "https://" + toaster.ID + "." + config.ToasterDomain + ":" + config.APIPort
		} else {
			toaster.RunLink = "https://" + toaster.ID + "." + config.ToasterDomain
		}
	}

	return toaster
}

func setToasterExeInfo(toaster *model.Toaster) error {
	dt := common.DumpToaterExeInfo(toaster)

	// the exe information dump contains the toaster's codeID and not its ID
	// this allows code updates which do not require runner cache invalidations
	err := redisdb.GetClient().Set(context.Background(), "exeinfo_"+toaster.ID, dt, 0).Err()
	if err != nil {
		return fmt.Errorf("could not save exe info in redis: %v", err)
	}

	var linkedSubs []*model.SubDomain
	linkedSubs, err = objectdb.Client.GetLinkedSubDomains(toaster.ID)
	if err != nil && err != objectdberror.ErrNotFound {
		return fmt.Errorf("could not get toaster linked subdmains: %v", err)
	}

	if len(linkedSubs) > 0 {
		pipe := redisdb.GetClient().Pipeline()
		for i := 0; i < len(linkedSubs); i++ {
			pipe.Set(context.Background(), "exeinfo_"+linkedSubs[i].Name, dt, 0)
		}
		_, err = pipe.Exec(context.Background())
		if err != nil {
			return fmt.Errorf("could not set toaster linked subdmains exe information: %v", err)
		}
	}

	return nil
}
