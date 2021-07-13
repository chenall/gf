package gspath

import (
	"github.com/gogf/gf/container/garray"
	"github.com/gogf/gf/os/gfile"
	"github.com/gogf/gf/os/gres"
	"github.com/gogf/gf/text/gstr"
)

// SearchWithRes 从 gres 和 文件系统 中查找指定文件,返回 filePath and resource
// 优先使用文件系统的文件
func SearchWithRes(searchPaths *garray.StrArray, fileName string, subDir string) (filePath string, resource *gres.File) {

	resourceTryFiles := []string{""}
	if subDir != "" {
		resourceTryFiles = append(resourceTryFiles, subDir)
	}
	searchPaths.RLockFunc(func(array []string) {
		for _, prefix := range array {
			prefix = gstr.TrimRight(prefix, `\/`)
			for _, v := range resourceTryFiles {
				if filePath, _ = Search(prefix+gfile.Separator+v, fileName); filePath != "" {
					return
				}
			}
		}
	})

	if filePath != "" {
		return
	}

	if !gres.IsEmpty() {
		resourceTryFiles := []string{"", "/"}
		if subDir != "" {
			resourceTryFiles = append(resourceTryFiles, subDir+"/", subDir, "/"+subDir, "/"+subDir+"/")
		}
		for _, v := range resourceTryFiles {
			if resource = gres.Get(v + fileName); resource != nil {
				filePath = resource.Name()
				return
			}
		}
		searchPaths.RLockFunc(func(array []string) {
			for _, prefix := range array {
				for _, v := range resourceTryFiles {
					if resource = gres.Get(prefix + v + fileName); resource != nil {
						filePath = resource.Name()
						return
					}
				}
			}
		})
	}
	return
}
