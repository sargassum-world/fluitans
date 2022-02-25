package embeds

import (
	"fmt"
	"io/fs"

	"github.com/pkg/errors"

	"github.com/sargassum-eco/fluitans/pkg/framework/fsutil"
	"github.com/sargassum-eco/fluitans/pkg/framework/template"
)

func IdentifyModuleNonpageFiles(templates fs.FS) (map[string][]string, error) {
	modules, err := fsutil.ListDirectories(templates, template.FilterModule)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't list template modules")
	}

	moduleFiles := make(map[string][]string)
	for _, module := range modules {
		var subfs fs.FS
		if module == "" {
			subfs = templates
		} else {
			subfs, err = fs.Sub(templates, module)
		}
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("couldn't list template module %s", module))
		}
		moduleSubfiles, err := fsutil.ListFiles(subfs, template.FilterNonpage)
		moduleFiles[module] = make([]string, len(moduleSubfiles))
		for i, subfile := range moduleSubfiles {
			if module == "" {
				moduleFiles[module][i] = subfile
			} else {
				moduleFiles[module][i] = fmt.Sprintf("%s/%s", module, subfile)
			}
		}
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("couldn't list template files in module %s", module))
		}
	}
	return moduleFiles, nil
}
