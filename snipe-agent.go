package main

// Import all the things
import (
  "fmt"
  "time"
  "encoding/json"
  "os"
  "os/exec"
  "net/http"
  "net/url"
  "io/ioutil"
  "strings"
  "flag"
  "runtime"
  "net"
  "log"
  "bytes"
)

// Set all the variables.
var NetworkHost string = "https://google.com" //website to test if network is up and to get preferred local IP address.
var SnipeHost string = ""
var SnipeKey string = ""
var UpdateFrequency = 15 // In Minutes
var SnipeID = 0 // This should always be set to 0 as snipe-agent will change it once the SN lookup succeeds. 
var BuildVersion string = "v0.1.0"
var StatusID = 1 // The Snipe Status ID that assets running the agent should be set to.

// Define all the Structs
// SnipeResults is a base structure for parsing returned search results.
type SnipeResults struct{
    Total       int             `json:"total"`
    AssetList   []AssetProfile  `json:"rows"`
}

// AssetProfile is an object style representation of returned Snipe-IT data.
type AssetProfile struct{
    Id        int         `json:"id"`
    Name      string      `json:"name"`
    AssetTag  string      `json:"asset_tag"`
    Serial    string      `json:"serial"`
    Notes     string      `json:"notes"`
}

// SnipeUpdatePayload is an object style that represents the patch Snipe-IT data.
type SnipeUpdatePayload struct{
    Name                    string    `json:"name"`
    Status_id               int       `json:"status_id"`
}

// Create Functions
func GetExternalIP() string {
  switch os := runtime.GOOS; os {
    case "windows":
      // Powershell invoke Rest Method to get the IP.
      cmd := exec.Command("powershell", "Invoke-RestMethod", "http://ipinfo.io/json", "|", "Select", "-exp", "ip")
      result, err := cmd.Output()
      if err != nil {
        fmt.Println(err)
        // Return err on error so we don't try to update.
        return "err"
      }
      // The result, by default contains carriage returns we need to remove:
      return strings.Replace(string(result), "\r\n", "", -1)
    // Linux command for the external IP.
    case "linux":
      // Powershell invoke Rest Method to get the IP.
      cmd := exec.Command("dig", "@resolver1.opendns.com", "ANY", "myip.opendns.com", "+short")
      result, err := cmd.Output()
      if err != nil {
        fmt.Println(err)
        // Return err on error so we don't try to update.
        return "err"
      }
      // The result, by default contains carriage returns we need to remove:
      return strings.Replace(string(result), "\r\n", "", -1)
    // macOS
    case "darwin":
      // Powershell invoke Rest Method to get the IP.
      cmd := exec.Command("dig", "@resolver1.opendns.com", "ANY", "myip.opendns.com", "+short")
      result, err := cmd.Output()
      if err != nil {
        fmt.Println(err)
        // Return err on error so we don't try to update.
        return "err"
      }
      // The result, by default contains carriage returns we need to remove:
      return strings.Replace(string(result), "\r\n", "", -1)
    case "default":
      fmt.Print("OS not Supported")
      return "err"
  }
  return "err"
}

func GetSerialNumber() string {
  switch os := runtime.GOOS; os {
    case "windows":
      // Run a powershell command to grab the Serial Number
      cmd := exec.Command("powershell", "gwmi", "win32_bios", "|", "Select-Object", "-ExpandProperty", "SerialNumber")
      sn, err := cmd.Output()
      if err != nil {
        fmt.Println(err)
        // Return err on error so we don't try to update.
        return "err"
      }
      // The serial number is byte encoded so when we return it, convert it to a string.
      return string(sn)
    default:
      fmt.Print("OS not Supported")
      return "err"
  }
}

// universal get hostname
func GetHostName() string{
  name, err := os.Hostname()
  if err != nil {
    fmt.Println(err)
    return "err"
  }
  return string(name)
}

// Get preferred outbound ip of this machine
func GetPreferredLocalIP() string {
    var ConnectionString string = ""
    if strings.HasPrefix(NetworkHost, "https://"){
      //remove the https prefix and form the string for the connection
      ConnectionString = strings.Replace(NetworkHost, "https://", "", -1) + ":https"
    }
    if strings.HasPrefix(NetworkHost, "http://"){
      //remove the http prefix and form the string for the connection
      ConnectionString = strings.Replace(NetworkHost, "http://", "", -1) + ":http"
    }
    conn, err := net.Dial("tcp", ConnectionString)
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    localAddr := conn.LocalAddr().(*net.TCPAddr)

    return localAddr.IP.String()
}

func GetCurrentUser() string {
  switch os := runtime.GOOS; os {
    // Make a case for Windows computers.
    case "windows":
      // Run a powershell command to grab the Serial Number
      cmd := exec.Command("powershell", "gwmi", "win32_process", "-f", `'Name="explorer.exe"'`, "|", "%", "getowner", "|", "%", "user" )
      result, err := cmd.Output()
      if err != nil {
        fmt.Println(err)
        // Return err on error so we don't try to update.
        return "err"
      }
      // The result, by default contains carriage returns we need to remove:
      return strings.Replace(string(result), "\r\n", "", -1)

    // Make a case for macOS computers.
    case "darwin":
      // Run a command to get the username that currently owns the console.
      cmd:= exec.Command("stat", "-f", "'%Su'", "/dev/console" )
      result, err := cmd.Output()
      if err != nil {
        fmt.Println(err)
        // Return err on error so we don't try to update.
        return "err"
      }
      // The result, by default contains carriage returns we need to remove:
      return strings.Replace(string(result), "\r\n", "", -1)

    // There is no case for the remaining OS types, so return an error.
    default:
      fmt.Print("OS not Supported")
      return "err"
  }
}

func FindSnipeID() int {
  // Get the SerialNumber so we can perform a search in Snipe-IT
  SerialNumber := GetSerialNumber()
  if SerialNumber == "err"{
    // We've Errored, so return 0
    return 0
  }

  // Perform lookup based off SN
  // Create the Web address we need.
  EncodedSN := &url.URL{Path: SerialNumber}
  // Remove Carriage and newLine Returns from encoded string.
  FixedEncodedSN := strings.Replace(EncodedSN.String(), "%0D%0A", "", -1)
  web := SnipeHost + "/api/v1/hardware/byserial/" + FixedEncodedSN

  // Set up the request and headers.
  req, err := http.NewRequest("GET", web, nil)
  req.Header.Add("Authorization", "Bearer " + SnipeKey)
  req.Header.Add("Accept", "application/json")

  // Send the request with http client
  client := &http.Client{}
  response, err := client.Do(req)
  if err != nil {
    fmt.Println("Got an error:")
    fmt.Println(err)
    // We failed, so return 0
    return 0
  }
  if response.StatusCode != 200 {
    // We didn't get a 200 response so return a failure.
    fmt.Println("Received an invalid response: ", response.StatusCode)
    // We failed, so return 0
    return 0
  }
  body, _ := ioutil.ReadAll(response.Body)
  var sniperesults SnipeResults
  json.Unmarshal(body,&sniperesults)
  if sniperesults.Total != 1 {
    // We got too many or few results, so return 0 for our loop.
    return 0
  }
  // We don't need to check for a null value, since INT returns 0 in that case. So just shipit.
  return sniperesults.AssetList[0].Id
}

// Recieves updated payload and ships it to snipe
func PatchToSnipe(assetPayload SnipeUpdatePayload) bool{
  // Create he URI to patch the asset.
  web := fmt.Sprintf("%s/api/v1/hardware/%d", SnipeHost, SnipeID)

  // Marshal updated payload becuase api call needs a marshalled json
  jsonPayload, err := json.Marshal(assetPayload)
  if err != nil {
    fmt.Println("Got an error:")
    fmt.Println(err)
    // We failed, so return false
    return false
  }
  // Create the body from the JSON.
  body := []byte(string(jsonPayload))

  // Set up the request and headers.
  req, err := http.NewRequest("PATCH", web, bytes.NewBuffer(body))
  if err != nil {
    fmt.Println("Got an error:")
    fmt.Println(err)
    // We failed, so return false
    return false
  }
  req.Header.Add("Authorization", "Bearer " + SnipeKey)
  req.Header.Add("Accept", "application/json")
  req.Header.Add("Content-type", "application/json")

  // Send the request with http client
  client := &http.Client{}
  response, err := client.Do(req)
  if err != nil {
    fmt.Println("Got an error:")
    fmt.Println(err)
    // We failed, so return false
    return false
  }
  defer response.Body.Close()

  if response.StatusCode != 200 {
    // We didn't get a 200 response so return a failure.
    fmt.Println("Received an invalid response: ", response.StatusCode)
    fmt.Println("Response Body: ", response.Body)
    // We failed, so return false
    return false
  }

  return true
}

// Recieves a struct and returns a updated struct
func PopulatePayload(assetInfo SnipeUpdatePayload) SnipeUpdatePayload{
  assetInfo.Name = GetHostName()
  assetInfo.Status_id = StatusID
  return assetInfo
}

func CheckWebHost(web string) bool{
  response, errors := http.Get(web)
  if errors != nil {
    // Fail, because there was an error.
    return false
  }
  if response.StatusCode == 200 {
    return true
  }
  // We didn't get a 200 response so return a failure.
  return false
}

// Main
func main() {
  // Set up Runtime flags.
  version := flag.Bool("version", false, "Report the version number and quit.")
  flag.Parse()
  // If the --version flag is present, report the BuildVersion and exit.
  if *version {
    fmt.Println(BuildVersion)
    os.Exit(0)
  }

  // Since this will be a service, set up a loop that repeats over an interval
  for {

    // Run some checks.
    // Try to connect to a stable network server.
    if CheckWebHost(NetworkHost) != true {
      fmt.Printf("%s could not be reached. Will retry at the next checkin interval.\n", NetworkHost)
      time.Sleep( time.Duration(UpdateFrequency) * time.Minute)
      // Continue so we don't run any more code and start the loop over again.
      continue
    }
    fmt.Println("Network seems up.")

    // Try to contact SnipeHost - Cycle on a failure with UpdateFrequency.
    if CheckWebHost(SnipeHost) != true {
      fmt.Printf("%s could not be reached. Will retry at the next checkin interval.\n", SnipeHost)
      time.Sleep( time.Duration(UpdateFrequency) * time.Minute)
      // Continue so we don't run any more code and start the loop over again.
      continue
    }
    fmt.Println("Snipe-IT instance seems up.")

    // If we don't know the SnipeID, look it up. - This way limits the number of API calls we need to make.
    if SnipeID < 1 {
      SnipeID = FindSnipeID()
      if SnipeID < 1 {
        fmt.Printf("The Snipe ID could not be found. Will retry at the next checkin interval.\n")
        time.Sleep( time.Duration(UpdateFrequency) * time.Minute)
        // Continue so we don't run any more code and start the loop over again.
        continue
      }
    }

    // Create a new blank SnipeUpdatePayload
    var assetInfo SnipeUpdatePayload

    // Pass blank payload into populate function which returns the current data.
    assetPayload := PopulatePayload(assetInfo)

    // Pass filled struct into function to update snipe
    // Set the returned value to update so we can quit if there was an error.
    update := PatchToSnipe(assetPayload)

    // Check if the update failed or not so we can fail the service if needed.
    if update != true {
      // We've failed, raise a system exit.
      fmt.Println("Patching the snipe asset failed, which shouldn't have occured. There is likely something wrong with your build. Exiting.")
      os.Exit(1)
    }
    // Sleep interval to delay updating again
    time.Sleep( time.Duration(UpdateFrequency) * time.Minute)
  }
}
