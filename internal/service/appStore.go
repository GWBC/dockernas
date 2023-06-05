package service

import (
	"dockernas/internal/backend/docker"
	"dockernas/internal/config"
	"dockernas/internal/models"
	"dockernas/internal/utils"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
)

var appMap map[string]*models.App
var appMap2 map[string]*models.App

func GetApps() []models.App {
	apps := []models.App{}

	dir1, err1 := ioutil.ReadDir("./apps")
	if err1 != nil {
		log.Println("list dir error", err1)
	} else {
		for _, fi1 := range dir1 {
			if fi1.IsDir() {
				dir2, err2 := ioutil.ReadDir(filepath.ToSlash(filepath.Join("./apps", fi1.Name())))
				if err2 != nil {
					log.Println("list dir error", err2)
				} else {
					for _, fi2 := range dir2 {
						if fi2.IsDir() {
							app := GetAppByNameAndPath(fi2.Name(), "./apps/"+fi1.Name()+"/"+fi2.Name())
							if app != nil {
								apps = append(apps, *app)
							}
						}
					}
				}
			}
		}
	}

	dir1, err1 = ioutil.ReadDir(config.GetExtraAppPath())
	if err1 != nil {
		log.Println("list dir error", err1)
	} else {
		for _, fi1 := range dir1 {
			if fi1.IsDir() {
				dir2, err2 := ioutil.ReadDir(filepath.ToSlash(filepath.Join(config.GetExtraAppPath(), fi1.Name())))
				if err2 != nil {
					log.Println("list dir error", err2)
				} else {
					for _, fi2 := range dir2 {
						if fi2.IsDir() {
							app := GetAppByNameAndPath(fi1.Name()+"/"+fi2.Name(), config.GetExtraAppPath()+"/"+fi1.Name()+"/"+fi2.Name())
							if app != nil {
								apps = append(apps, *app)
							}
						}
					}
				}
			}
		}
	}

	appMap = make(map[string]*models.App, 0)
	appMap2 = make(map[string]*models.App, 0)

	for _, app := range apps {
		tApp := app
		appMap[app.Name] = &tApp
		for _, v := range app.DockerVersions {
			index := strings.Index(v.ImageUrl, ":")
			if index < 0 {
				v.ImageUrl += ":latest"
			}
			appMap2[v.ImageUrl] = &tApp
		}
	}

	return apps
}

func GetAppByName(name string, flush bool) *models.App {
	if appMap == nil {
		GetApps()
	}

	app, ok := appMap[name]
	if ok {
		return GetAppByNameAndPath(app.Name, config.GetAbsolutePath(app.Path)) //get lastest data on disk
	}
	if !flush {
		return nil
	}

	GetApps()
	return GetAppByName(name, false)
}

func GetAppByImage(image string) (*models.App, *models.DockerTemplate) {
	if appMap2 == nil {
		GetApps()
	}

	app, ok := appMap2[image]
	if ok {
		app = GetAppByNameAndPath(app.Name, config.GetAbsolutePath(app.Path)) //get lastest data on disk
		var template *models.DockerTemplate = nil

		if app != nil {
			for _, v := range app.DockerVersions {
				index := strings.Index(v.ImageUrl, ":")
				if index < 0 {
					v.ImageUrl += ":latest"
				}

				if v.ImageUrl == image {
					template = &v
					break
				}
			}

			if template == nil {
				template = &app.DockerVersions[0]
			}
		}

		return app, template
	}

	return nil, nil
}

func GetAppByNameAndPath(name string, path string) *models.App {
	var app models.App
	app.IconUrl = "/api/icon?path=" + config.GetRelativePath(path) + "/icon.jpg"
	app.DockerVersions = getDockerTemplates(name, path+"/docker")
	if len(app.DockerVersions) == 0 {
		return nil
	}
	if utils.GetObjFromJsonFile(path+"/introduction.json", &app) == nil {
		return nil
	}
	app.Name = name
	app.Path = config.GetRelativePath(path)

	return &app
}

func getDockerTemplates(name string, path string) []models.DockerTemplate {
	var dockerTemplates []models.DockerTemplate

	dirs, err := ioutil.ReadDir(path)
	if err != nil {
		log.Println("list dir error", err)
		return dockerTemplates
	}

	for _, fi := range dirs {
		if fi.IsDir() {
			var dockerTemplate models.DockerTemplate
			dockerTemplate.Version = fi.Name()
			if utils.GetObjFromJsonFile(path+"/"+fi.Name()+"/template.json", &dockerTemplate) != nil {
				if dockerTemplate.OSList != "" &&
					!strings.Contains(dockerTemplate.OSList, docker.DetectRealSystem()) {
					continue
				}
				dockerTemplate.Path = config.GetRelativePath(path) + "/" + fi.Name()
				dockerTemplates = append(dockerTemplates, dockerTemplate)
			} else {
				log.Println("load template error for " + fi.Name() + " under " + path)
			}
		}
	}

	for i, v := range dockerTemplates {
		for j, dfs := range v.DfsVolume {
			if len(dfs.Value) == 0 || dfs.Value == "/" {
				dockerTemplates[i].DfsVolume[j].Value = filepath.ToSlash(filepath.Join("/", name+"_"+v.Version))
			}
		}
	}

	return dockerTemplates
}
