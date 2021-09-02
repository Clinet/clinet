package main

import (
	//"github.com/Clinet/clinet/bot/modules/core"
	//A convenient std plugin replacement written for me to reload plugins
	//"github.com/superwhiskers/uwu"

	//std necessities
	"io/ioutil"
)

var Modules []*Module

type Module struct {
//	Plugin *uwu.Plugin
}

func loadModules() {
	log.Trace("--- loadModules() ---")
	log.Warn("Modules are not supported yet!")

	/*wd, _ := os.Getwd()
	modulesDir := wd + "/modules"

	log.Debug("Searching for modules")
	recurseCall(modulesDir, func(path string) {
		if strings.HasSuffix(path, ".so") {
			log.Debug("Loading module (", path, ") into memory")

			plugin, err := uwu.LoadPlugin(path)
			if err != nil {
				log.Error("Error loading module (", path, ") into memory: ", err)
				return
			}

			log.Debug("Loading symbols from module (", path, ")")
			err = plugin.Load()
			if err != nil {
				log.Error("Error loading symbols from module (", path, "): ", err)
				plugin.Unload()
				return
			}

			log.Debug("Initializing module (", path, ")")
			err = plugin.Init()
			if err != nil {
				log.Error("Error initializing module (", path, "): ", err)
				plugin.Unload()
				return
			}

			log.Debug("Calling for metadata from module (", path, ")")
			voidMetadata, err := plugin.Call("GetMetadata")
			if err != nil {
				log.Error("Error getting metadata from module (", path, "): ", err)
				plugin.Unload()
				return
			}
			ptrMetadata := unsafe.Pointer(&voidMetadata)
			metadata := *(*modulecore.Metadata)(ptrMetadata)
			log.Debug("METADATA: ", metadata)

			plugin.Unload()
		}
	})*/
}

func recurseCall(path string, function func(path string)) error {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return err
	}
	for _, result := range files {
		if result.IsDir() {
			recurseCall(path + "/" + result.Name(), function)
		}

		go function(path + "/" + result.Name())
	}
	return nil
}