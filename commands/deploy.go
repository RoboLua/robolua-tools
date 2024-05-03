package commands

import (
	"os"
	"path/filepath"

	"github.com/charmbracelet/log"
	"github.com/melbahja/goph"
)

func fileExists(path string) bool {
	_, e := os.Stat(path)
	if e == nil {
		return true;
	}

	return false;
}

func recursiveUpload(client *goph.Client, folderPath string, shpath string) error {
	files, err := os.ReadDir(folderPath)

	if err != nil {
		log.Error("Couldn't read local directory", "err", err)
		return err
	}

	for _, file := range files {
		if file.IsDir() {

			_, err = client.Run("mkdir " + shpath + "/" + file.Name())

			if err != nil {
				log.Error("Couldn't create new remote directory", "path", shpath+"/"+file.Name(), "err", err)
				return err
			}

			recursiveUpload(client, folderPath+"\\"+file.Name(), shpath+"/"+file.Name())
		} else {
			err = client.Upload(folderPath+"\\"+file.Name(), shpath+"/"+file.Name())

			if err != nil {
				log.Error("Couldn't upload file", "local_path", folderPath+"\\"+file.Name(), "remote_path", shpath+"/"+file.Name(), "err", err)
				return err
			}
		}
	}

	return nil
}

func Deploy() {
	working_dir, err := os.Getwd()

	if err != nil {
		log.Fatal("Couldn't get working directory and -d was not specified...", "err", err)
	}

	if !fileExists(working_dir + "/src/main.lua") {
		log.Fatal("Couldn't find main.lua, make sure you're running in the root of the project directory.")
	}

	if !fileExists(working_dir + "/deploy") {
		os.Mkdir(working_dir+"/deploy", 04|02)
		os.WriteFile(working_dir+"/deploy/example.txt", []byte("Anything in this folder will be copied to the roborio at /home/lvuser/deploy\n"), 04|02)
	}

	log.Info("Connecting to roborio...")
	client, err := goph.New("admin", "172.22.11.2", goph.Password(""))

	if err != nil {
		log.Fatal("Couldn't connect to roborio, is the bot running?", "err", err)
	}

	_, err = client.Run("/usr/local/frc/bin/frcKillRobot.sh -t")

	if err != nil {
		log.Fatal("Couldn't kill the robot, deploying when the robot is unsafe.", err, err)
	}

	log.Debug("Killed the robot, deploying new code...")

	sftp_client, err := client.NewSftp()

	if err != nil {
		log.Fatal("Couldn't create sftp client", "err", err)
	}

	_, doesnt_exist := sftp_client.ReadDir("/home/lvuser/deploy")

	if doesnt_exist == nil {
		_, err = client.Run("rm -rf /home/lvuser/deploy")

		if err != nil {
			log.Fatal("Failed to remove old deploy directory", "err", err)
		}
	}

	log.Debug("Removed old deploy directory")

	client.Run("mkdir /home/lvuser/deploy")
	err = recursiveUpload(client, working_dir+"\\deploy", "/home/lvuser/deploy")

	if err != nil {
		log.Error("Failed to copy deploy directory to roborio, stopping deployment to precent unsafe launch.", "err", err)
		log.Warn("Old robot code dependent on the deploy directory will fail, a re-deploy is not neccessary.")
		return
	}

	log.Info("Successfully cloned deploy directory to roborio, cloning robotcode...")

	_, err = client.Run("rm -rf /home/lvuser/src")

	if err != nil {
		log.Fatal("Failed to delete old source", "err", err)
	}
	
	client.Run("mkdir /home/lvuser/src")
	err = recursiveUpload(client, working_dir+"\\src", "/home/lvuser/src")

	if err != nil {
		log.Fatal("Failed to copy robot code to roborio, is the roborio running?", "err", err)
	}

	log.Info("Succesfully Uploaded source")
	_, robolua_not_installed := sftp_client.ReadDir("/usr/local/frc/robolua")


	log.Debug(robolua_not_installed)
	if robolua_not_installed != nil {

		log.Info("Robolua not installed, installing...")

		executable_location, err := os.Executable()

		if err != nil {
			log.Fatal("Couldn't get executable location", "err", err)
		}

		log.Debug(executable_location, filepath.Dir(executable_location))
		executable_dir, err := os.ReadDir(filepath.Dir(executable_location))

		if err != nil {
			log.Fatal("Couldn't get files associated with robolua, running robolua-tools verify", "err", err, "expected_path", executable_dir)
		}

		robolua_found := false
		for _, file := range executable_dir {
			if file.Name() == "robolua" {
				log.Info("Found robolua executable, copying to roborio...")
				err = client.Upload(filepath.Dir(executable_location)+"\\robolua", "/usr/local/frc/robolua")

				if err != nil {
					log.Fatal("Failed to upload robolua to roborio", "err", err)
				}

				robolua_found = true

				break
			}
		}

		if !robolua_found {
			log.Fatal("Couldn't find the robolua executable, run robolua-tools verify to make sure it's installed correctly.", "expected_path", executable_dir)
		}

		_, err = client.Run("chmod +x /usr/local/frc/robolua")

		if err != nil {
			log.Fatal("Failed to make robolua executable", "err", err)
		}

		log.Info("Successfully installed robolua!")
	}

	_, err = client.Run("echo '/usr/local/frc/robolua /home/lvuser/src/main.lua' > /home/lvuser/robotCommand")

	if err != nil {
		log.Fatal("Failed to write robot command to roborio, is the roborio running?", "err", err)
	}

	
	log.Info("Completed Deployment!")
}