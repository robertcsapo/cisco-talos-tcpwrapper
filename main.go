// cisco-talos-tcpwrapper (/etc/hosts.deny)
// Creator: @robertcsapo github / twitter

package main

import (
  "bufio"
  "fmt"
  "io/ioutil"
  "strings"
  "net/http"
  "os"
  "log"
  "time"
  "gopkg.in/alecthomas/kingpin.v2"
)

var (
  args        = kingpin.New(os.Args[0], "Cisco Talos tcpwrapper command-line to override settings")
  debugArg       = args.Flag("debug", "Enable debug mode").Bool()
  urlArg         = args.Flag("url", "Change URL for Cisco Talos IP Blacklist").String()
  sleepArg       = args.Flag("sleep", "Change sleep timer between downloads of the list").Int()
  helpArg        = args.HelpFlag.Short('h')
  versionArg     = args.Version("cisco talos tcpwrapper version 1.0")
)

func start() {

  // Settings
  path := "/etc/"
  cisco_talos_file := "cisco-talos-tcpwrapper/cisco-talos-ip-blacklist"
  url := "https://www.talosintelligence.com/documents/ip-blacklist"
  // Better log view in docker logs
  log.SetOutput(os.Stdout)

  // Toggle to enable Debug (true/false)
  debug := false
  if *debugArg == true {
    debug = true
    log.Printf("DEBUG: \tEnabled (%t)",debug)
  }

  // Change URL if urlArg is set
  if *urlArg != "" {
    url = *urlArg
  }

  // Time used for putting a timestamp on Cisco Talos File line 0
  currentTime := time.Now()
  log.Printf("START: \tDownloadning from %s",url)

  // Download the ip-blacklist from Cisco Talos
  req, _ := http.NewRequest("GET", url, nil)
  res, err := http.DefaultClient.Do(req)
  if err != nil {
    log.Fatalf("ERROR: \tCan't download Cisco Talos from %s (%s)\n",url, err)
  }
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
  if debug == true {
    fmt.Println(string(body))
  }
  if (res.StatusCode != 200) {
    if (res.StatusCode == 429) {
      log.Printf("ERROR: \tCisco Talos: %s. Waiting 10 minutes before exit",string(body))
      time.Sleep(time.Second * time.Duration(600))
      log.Fatal("QUIT: \tThrottle timer expired. Please restart.")
    } else {
      log.Fatalf("ERROR: \tCan't download Cisco Talos from %s (%s)\n",url, res.Status)
    }
  }

  // Create/Edit the cisco talos file (don't use full path as it's a container)
  file, err := os.Create(cisco_talos_file)
    if err != nil {
        log.Fatal("ERROR: \tCannot create/edit file", err)
    }
  defer file.Close()
  // Clean file incase of
  fmt.Fprintf(file, "")
  // Write to clean file
  fmt.Fprintf(file,"# UPDATED: %s\n", currentTime.Format("2006.01.02 15:04:05"))
  fmt.Fprintf(file, string(body))
  log.Printf("SUCCESS: \tDownloaded %s to %s%s\n",url,path,cisco_talos_file)
  // Open/Edit the hosts.deny
  filename := "hosts.deny"
  content, err := ioutil.ReadFile(filename)
  if err != nil {
      log.Fatal("ERROR: \tCannot create/edit file", err)
  }

  // If we catch cisco talos file in hosts.deny
  lines := strings.Split(string(content), "\n")
  catch := false
  // If previous configuration is found in hosts.deny. Report config line
  content_line := 0
  for _, content := range lines {
    content_line++
    if strings.HasPrefix(content, "#") != true && content != "" && strings.Contains(content, cisco_talos_file) {
      if strings.Contains(content, path+cisco_talos_file) {
        log.Printf("SUCCESS: \tFound %s in %shosts.deny",path+cisco_talos_file,path)
          if debug == true {
            log.Printf("DEBUG: \tData to put in hosts.deny")
            log.Printf("DEBUG: \thosts.deny -> %s\n",content)
            }
        // we found cisco tales file
        catch = true
      } else {
        log.Printf("ISSUE: \tSeems to be another %s in %shosts.deny\n",cisco_talos_file,path)
        if debug == true {
          log.Printf("DEBUG: \t%shosts.deny:%v: %s",path,content_line,content)
        }
      }
    }
  }

  // Didn't find cisco talos file in hosts.deny. Attempts to add it to hosts.deny file
  if catch == false {
    log.Printf("ISSUE: \tCouldn't find %s in %shosts.deny\n",cisco_talos_file,path)
    file, err = os.OpenFile("hosts.deny", os.O_APPEND|os.O_WRONLY, 0644)
      if err != nil {
          log.Fatal("ERROR: \tCannot open/edit file", err)
      }
    defer file.Close()
    // Append the hosts.deny file with Cisco Talos file
    w := bufio.NewWriter(file)
    full_path := []string{path, cisco_talos_file}
    // Logger will report Cisco Talos blocks in to syslog
    write_string := "ALL: "+strings.Join(full_path, "")+": spawn /usr/bin/logger DENY %h blocked due to Cisco Talos\n"
    w.WriteString(write_string)
    w.Flush()
    log.Printf("SUCCESS: \tAdded %s to %shosts.deny\n",cisco_talos_file,path)
  }
  //Display Args when using Debug
  if debug == true {
    fmt.Printf("\n\n")
    log.Printf("DEBUG VARS:\n")
    log.Printf("PATH %s\n",path)
    log.Printf("CISCO_TALOS_FILE %s\n",cisco_talos_file)
    log.Printf("FULL PWD %s%s\n",path,cisco_talos_file)
    log.Printf("URL %s\n",url)
    fmt.Printf("\n\n")
    log.Printf("DEBUG ARGs:\n")
    if *urlArg == "" {
      *urlArg = url+" (default)"
    }
    log.Printf("URL: %s\n",*urlArg)
    if *sleepArg == 0 {
      *sleepArg = 3600
      log.Printf("SLEEP: %v (default)\n",*sleepArg)
    } else {
      log.Printf("SLEEP: %v\n",*sleepArg)
    }
    log.Printf("DEBUG: END\n\n\n\n\n")
  }
}

// Start script and loop
func main() {
  //Prints args
  kingpin.MustParse(args.Parse(os.Args[1:]))
  for {
    start()
    seconds := 3600
    if *sleepArg != 0 {
      seconds = *sleepArg
    }
    // Wait X (sleeptime var) seconds before running again. Use --sleep arg to override if needed
    time.Sleep(time.Second * time.Duration(seconds))
  }
}
