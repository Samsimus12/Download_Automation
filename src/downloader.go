package main

import (
   cdp "github.com/knq/chromedp"
   "context"
   "github.com/fatih/color"
   "time"
   "strconv"
   "strings"
   "regexp"
   "os"
   "io"
   "fmt"
	"io/ioutil"
)

var c *cdp.CDP

var rootDirectory = "/Users/Desired-Save-Folder-Location"
var rootFolder = "Zapsplat"
var downloadsFolder = "/User/Downloads-Folder"
var totalDownloads int

func crawl(ctxt context.Context, category, sounds string) {
   time.Sleep(2 * time.Second)
   color.Yellow("BEGIN DOWNLOADING %v LIBRARY", category)
   var location, result, evalJS string
   var pageLength, remainder, categoryLength, failCount int
   page := "https://www.zapsplat.com/sound-effect-category/" + category + "/?pageCustom=100"
   count := 1
   soundsCount, _ := strconv.Atoi(sounds)
   if soundsCount > 100 {
      pageLength = (soundsCount / 100) + 1
      categoryLength = 200
      remainder = soundsCount % 100 * 2
   } else {
      pageLength = 1
      categoryLength = soundsCount * 2
   }

   c.Run(ctxt, cdp.Tasks{cdp.Navigate(page)})

   for i := 1; i <= pageLength; i++ {
      if i == pageLength && pageLength > 1 {
         categoryLength = remainder
      }
      for j := 0; j < categoryLength; j++ {
         if j % 2 != 0 {
            c.Run(ctxt, cdp.Location(&location))

            if !strings.Contains(location, category) && failCount < 3 {
               color.Red("ERROR WITH DOWNLOAD!!")
               failCount++
               c.Run(ctxt, cdp.Tasks{cdp.Navigate("https://www.zapsplat.com/sound-effect-category/" + category + "/page/" + strconv.Itoa(i) + "/?pageCustom=100")})
               color.Yellow("NAVIGATING BACK TO PAGE #%v", i)
               time.Sleep(1 * time.Second)
            } else {
            	totalDownloads++
			}
			//if strings.Contains(category, "Alligators") || categoryLength < 10 {
			//	time.Sleep(1 * time.Second)
			//}
            time.Sleep(500 * time.Millisecond)
            if waitOnDownloadQueue() {
				evalJS = `
				   var nodes = document.querySelectorAll('a.uk-button.uk-button-primary.downloadbutton');
				   nodes[` + strconv.Itoa(j) + `].click();
				`
				c.Run(ctxt, cdp.Tasks{cdp.Evaluate(evalJS, &result)})
				color.Cyan("DOWNLOADED %v SOUND #%v", category, count)
				count++
			}
         }
      }
      if i != pageLength {
         color.Yellow("NAVIGATING TO PAGE #%v", i+1)
         failCount = 0
         c.Run(ctxt, cdp.Tasks{cdp.Navigate("https://www.zapsplat.com/sound-effect-category/" + category + "/page/" + strconv.Itoa(i+1) + "/?pageCustom=100")})
      }
   }
}

func getCategoryList(ctxt context.Context, catContainerPath, catPath string) (categoryNames, categoriesCount, subCategoryCount []string) {
   var firstCatCount string
   var cat string

   time.Sleep(2 * time.Second)

   color.Cyan("ATTEMPTING TO ACQUIRE SUBCATEGORY LIST")

   c.Run(ctxt, cdp.Tasks{cdp.OuterHTML(catContainerPath, &firstCatCount, cdp.ByQuery)})
   //color.Yellow("LI %v", firstCatCount)
   reg, _ := regexp.Compile("\\([0-9]+\\)")
   fcc := reg.FindAllString(firstCatCount, -1)
   color.Cyan("SUBCATEGORY DOWNLOAD SOUNDS: %v", fcc)

   for i := 1; i <= len(fcc); i++ {
      time.Sleep(30 * time.Millisecond)
      c.Run(ctxt, cdp.Tasks{cdp.Text(catPath + ":nth-of-type(" + strconv.Itoa(i) + ")", &cat, cdp.ByQuery)})

      ns := strings.Replace(fcc[i-1], "(", "", -1)
      ns = strings.Replace(ns, ")", "", -1)
      ns = strings.Replace(ns, " ", "", -1)
      subCategoryCount = append(subCategoryCount, ns)

      reg2, _ := regexp.Compile("([a-zA-Z ]+)")
      catArray := reg2.FindAllString(cat, 1)
      cat = strings.Replace(catArray[0], "\n", "", -1)
      cat = strings.Replace(cat, " ", "-", -1)
      cat = cat[:len(cat)-1]
      categoryNames = append(categoryNames, cat)
   }

   categoriesCount = fcc
   //color.Yellow("CATNAMES %v CATCOUNT %v SUBCATCOUNT %v", categoryNames, categoriesCount, subCategoryCount)

   return categoryNames, categoriesCount, subCategoryCount
}

func copy_folder(source string, dest string) (err error) {

   sourceinfo, err := os.Stat(source)
   if err != nil {
      return err
   }

   err = os.MkdirAll(dest, sourceinfo.Mode())
   if err != nil {
      return err
   }

   directory, _ := os.Open(source)

   objects, err := directory.Readdir(-1)

   for _, obj := range objects {

      sourcefilepointer := source + "/" + obj.Name()

      destinationfilepointer := dest + "/" + obj.Name()

      if obj.IsDir() {
         err = copy_folder(sourcefilepointer, destinationfilepointer)
         if err != nil {
            fmt.Println(err)
         }
      } else {
         err = copy_file(sourcefilepointer, destinationfilepointer)
         if err != nil {
            fmt.Println(err)
         }
      }

   }
   return
}

func copy_file(source string, dest string) (err error) {
   sourcefile, err := os.Open(source)
   if err != nil {
      return err
   }

   defer sourcefile.Close()

   destfile, err := os.Create(dest)
   if err != nil {
      return err
   }

   defer destfile.Close()

   _, err = io.Copy(destfile, sourcefile)
   if err == nil {
      sourceinfo, err := os.Stat(source)
      if err != nil {
         err = os.Chmod(dest, sourceinfo.Mode())
      }

   }

   return
}

func ensureDownloadFinished() bool {
	color.Yellow("--------------- FINALIZING DOWNLOADS ---------------")
	for {
		time.Sleep(10 * time.Millisecond)
	    files, err := ioutil.ReadDir("/User/Downloads-Folder")
        if err != nil {
            color.Red("ERROR: %v", err)
        }
		for i := 0; i < len(files); i++ {
			if strings.Contains(files[i].Name(), "crdownload") {
			    break
			}
			if i == len(files) - 1 {
				//SAM MAYBE PUT COPYFOLDER() RIGHT HERE
				color.Green("--------------- DOWNLOAD FINISHED ---------------")
				color.Yellow("--------------- TRANSFERRING SOUNDS TO NEW FOLDER ---------------")
				return true
			}
		}
	}
    return true
}

func waitOnDownloadQueue() bool {
	color.Yellow("WAITING ON DOWNLOADS QUEUE")
	var queueCount int
	for {
		queueCount = 0
		time.Sleep(10 * time.Millisecond)
	    files, err := ioutil.ReadDir(downloadsFolder)
        if err != nil {
            color.Red("ERROR: %v", err)
        }
        for i := 0; i < len(files); i++ {
        	if strings.Contains(files[i].Name(), "crdownload") {
        		queueCount++
			}
		}
		if queueCount <= 6 {
			color.Green("%v DOWNLOADS IN QUEUE. PROCEED", queueCount)
			return true
		}
	}
    return true
}

func main() {
   startTime := time.Now()
   //os.Chdir(rootDirectory)
   //wd, _ := os.Getwd()
   //color.Magenta("CURRENT DIRECTORY: %v", wd)
   ctxt, cancel := context.WithCancel(context.Background())
   c, _ = cdp.New(ctxt)
   defer cancel()

   color.Yellow("--------------- LOGGING IN ---------------")
   c.Run(ctxt, cdp.Tasks{cdp.Navigate("https://www.zapsplat.com/login/"), cdp.SendKeys("input#user_login", "userEmail@example.com", cdp.ByQuery),
      cdp.SendKeys("input#user_pass", "userPassword", cdp.ByQuery), cdp.Click("#wppb-submit", cdp.ByQuery)})

   time.Sleep(5 * time.Second)
   c.Run(ctxt, cdp.Tasks{cdp.Navigate("https://www.zapsplat.com/sound-effect-categories")})
   var categoriesCount []string
   var categories []string
   var subCategoryNames []string
   var subCategoryCount []string
   var subCategorySounds []string
   var cat string
   var catCount string

   for i := 1; i <= 27; i++ {
      time.Sleep(30 * time.Millisecond)
      c.Run(ctxt, cdp.Tasks{cdp.Text("#post-content > div > ul > li:nth-of-type(" + strconv.Itoa(i) + ")", &catCount,  cdp.ByQuery),
                       cdp.Text("#text-9 > div > div > ul > li:nth-of-type(" + strconv.Itoa(i) + ")", &cat, cdp.ByQuery)})
      reg, _ := regexp.Compile("\\([0-9]+\\)")
      newString := reg.FindAllString(catCount, 1)
      ns := strings.Replace(newString[0], "(", "", -1)
      ns = strings.Replace(ns, ")", "", -1)
      categoriesCount = append(categoriesCount, ns)
      categories = append(categories, cat)
   }

   for i := 0; i < len(categories) - 1; i++ {
      formatCategory := strings.Replace(categories[i], "\n", "", -1)
      formatCategory = strings.Replace(formatCategory, " ", "-", -1)
      switch formatCategory {
	  		case "Explosions":
      			formatCategory = "Explosions-And-Fireworks"
		    case "Vehicles":
				formatCategory = "Vehicles-And-Transport"
		    case "Warfare":
		    	formatCategory = "War-And-Weapons"
	  }

      color.Cyan("MAIN CATEGORY: %v WITH %v SOUNDS", formatCategory, categoriesCount[i])
      page := "https://www.zapsplat.com/sound-effect-category/" + formatCategory
      c.Run(ctxt, cdp.Tasks{cdp.Navigate(page)})
      subCategoryNames, subCategoryCount, subCategorySounds = getCategoryList(ctxt, ".sub-cat-list", ".sub-cat-list > li")
      //color.Magenta("LOOPING THROUGH SUBCATEGORYCOUNT %v", subCategorySounds)

      for j := 0; j < len(subCategoryCount); j++ {
         color.Cyan("SUBCATEGORY: %v WITH %v SOUNDS", subCategoryNames[j], subCategorySounds[j])
         crawl(ctxt, subCategoryNames[j], subCategorySounds[j])
         os.Mkdir(formatCategory, 0777)
         os.Chdir(rootDirectory + "/" + formatCategory)
         os.Mkdir(subCategoryNames[j], 0777)
         //if ensureDownloadFinished() {
			 copy_folder(downloadsFolder, rootDirectory + "/" + formatCategory + "/" + subCategoryNames[j])
			 //if ensureDownloadsTransferred(formatCategory, subCategoryNames[j]) {
				 os.Chdir(rootDirectory)
				 if formatCategory == "Music" {
				 	time.Sleep(1 * time.Minute)
				 } else {
					 time.Sleep(30 * time.Second)
				 }
				 os.RemoveAll(downloadsFolder)
			 //}
		 //}
         time.Sleep(1 * time.Second)
      }
   }
   endTime := time.Now()
   elapsed := endTime.Sub(startTime)
   color.Green("--------------- TOTAL DOWNLOADS: %v ---------------", totalDownloads)
   color.Yellow("--------------- DOWNLOAD FINISH TIME: %v ---------------", elapsed)

   c.Shutdown(ctxt)
}