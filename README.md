# snipe-agent
Snipe Agent is written in Go Programming Language. The current goal for this
project is to have a functioning agent that works in Windows. The project may be
expanded to other operating systems at a later date.


***
## Configuration
To configure this program for your own use, you'll need to edit some of the base
code for your setup. Many, if not all, of the settings will be in the *Build*
section below, but not all are able to be set this way. It's importanty to note
that only variables that are Strings, (not INT or other types) can be set
through build variables. To configure other types, you'll need to edit the
variable section near the top of the `snipe-agent.go` file.

#### Updating Custom Snipe-IT Fields:
To enable the agent to update to a custom field, you'll need to find the name of
the custom field. The agent uses a PATCH call, and snipe has very good
documentation on using custom fields with their API:
https://snipe-it.readme.io/reference#hardware-partial-update
After you find what your custom field is, you'll need to use that field to add
it as a key to the payload. (see below)

#### Adding Key Value Pairs to the Patch Payload:
Adding new fields, such as a custom field, requires that you updated the
SnipeUpdatePayload struct. If we wanted the agent to update a field called
*_snipeit_custom_name_1* we would need to add it to the struct like so:

```
type SnipeUpdatePayload struct{
    Name          string    `json:"name"`
    Status_id     int       `json:"status_id"`
    CustomName    string    `json:"_snipeit_custom_name_1"`
}
```
Note that as you add the keyname to the struct, you've capitalized the first
letter of the key so that it can be exported where it's listed as `CustomName`
in the example above. Then you add the json key `_snipeit_custom_name_1`
in the final column so the name exists when the JSON is marshaled.

After you've added it there, you'll need to also update the Populate_Payload
function to set the values to the new key you added in the struct.
```
func PopulatePayload(assetInfo SnipeUpdatePayload) SnipeUpdatePayload{
  assetInfo.Name = get_host_name()
  assetInfo.Status_id = StatusID
  assetInfo.CustomName = MyVar
  return assetInfo
}
```
We recommend setting a static variable that is injected at build, or writing a
new function to obtain the data you need if it's not already available in go.

#### Available Functions That Provide Custom Data:
- `GetHostName()` - Returns the local computer's hostname.
- `GetCurrentUser()` - Returns the name of the current user.
- `GetSerialNumber()` - Returns the local device's serial number.
- `GetExternalIP()` - Returns external IP as a string.
- `GetPreferredLocalIP()` - Returns the local IPv4 address of the preferred network interface as a string.

---
## Build
With go, you're able to compile an executable that can be redistributed with the
required modules and libraries so you do not need them installed or
redistributed separately. This will become our main approach as we get ready to
deploy snipe-agent.

#### Injecting Build Variables:
With Go, we have the option to overwrite the variables by injecting them when we
use `go build` with the `-ldflags` option. For our purposes, this will generally
suffice to setting most of our variables by using this method. For example, if
we wanted the agent to have a custom version number, we can set that at compile
time with `go build -ldflags "-X 'main.BuildVersion=v0.1.0-200'"`. We can stack
these settings togethe. If we also wanted to change the Snipe instance, we can
set both of them together like:
`go build -ldflags "-X 'main.BuildVersion=v0.1.0-200' -X 'main.SnipeHost=https://develop.snipeitapp.com'"`

#### Building an .exe from snipe-agent.go:
Go does have and option to cross compile from other operating systems.
Typically, this would look like:
`GOOS=windows GOARCH=386 go build -o hello.exe hello.go`
So in using the example above from the project root, it would be:
```
GOOS=windows GOARCH=386 go build -ldflags "-X 'main.UpdateFrequency=45' -X 'main.SnipeHost=https://develop.snipeitapp.com'" -o snipe-agent.exe snipe-agent.go
```
You wouldn't need to set the first portion if you're building on Windows
natively, and another example may look something like this:
```
go build -ldflags "-X 'main.SnipeKey=<supersecretkeyhere>' -X 'main.BuildVersion=v0.1.0-2' -X 'main.SnipeHost=https://develop.snipeitapp.com'" -o snipe-agent.exe snipe-agent.go
```
#### Required Build Variables:
You will be required to set the following variables with `-ldflags` with in
order to have the agent function as intended:
```
main.SnipeKey
main.SnipeHost
```
Please see the variables that are set in the top of snipe-agent.go for a
complete list of things that can be set through Build Variables. Remember that
you will need to set them with `main.VariableName` and that they can only be
variables that are strings. If you do not wish to use `-ldflags` you can set
them directly in the `snipe-agent.go` file.


---
## Installation Guide
The first thing you'll need to do is build your own version of the snipe-agent.
After that, the fastest way to deploy the agent is to register the created
`snipe-agent.exe` as a windows service. Snipe Agent doesn't currently work as a
native Windows Service that can be managed by SCM. You can get around this by
registering the service with `cmd.exe` as the main process (which also doesn't
truly register as a windows service) and have it spawn an instance of
`snipe-agent.exe`. When Windows shuts down the instance of CMD.exe, thinking the
service failed, snipe-agent.exe will still be running. It would be better to
install with Windows Resource Kit, or some other 3rd party installer, but it
still is a valid way of getting the agent to run. You can configure snipe-agent
to run this way by using something like:
```
sc.exe create SnipeAgent binpath= "cmd.exe C:\Path\To\Local\snipe-agent.exe" type= own start= auto
```
`sc.exe delete SnipeAgent` would delete the service from the local machine.

#### Using SRVANY.EXE and the included install.ps1 script:
Alternatively, we've included `install.ps1` and `uninstall.ps1` to use in
conjunction with `srvany.exe` from the [Windows Resource Kit](https://www.microsoft.com/en-us/download/details.aspx?id=17657).
To deploy snipe-agent with the included powershell scripts, download and
extract the `srvany.exe` from the Windows Resource Kit and create a source
folder with the `install.ps1`, `uninstall.ps1`, `srvany.exe`, and your compiled
`snipe-agent.exe` in a source folder. That folder should be the working folder
for powershell when you start the install. The script will install `srvany`
and `snipe-agent.exe` to `%WINDIR%` by default. You can change those locations
as needed by editing the script. It registers your program as a service and also
creates a version property in the service registry key. If you are deploying
`snipe-agent` with SCCM, your install command would be:
`powershell.exe -ExecutionPolicy Bypass -File install.ps1` and you can set the
detection rule to find `hklm:\SYSTEM\CurrentControlSet\Services\SnipeAgent` as
the key and `Version` as the Value, The Type would be `String` and if you want
to specify the value of the string, it will match the output of your compiled
program: `snipe-agent.exe -version`

We strongly recommend running the compiled executable on it's own prior to
installation to test for any errors that might occur. It's also useful to call
the executable directly as a troubleshooting step.


***
## Getting Help
Unfortunately, the maintainers of the code base are not able to provide end-user
support for program. When creating issues against this repository, please keep
that in mind. If you need help, or have questions, there is a vibrant go
community on [freenode.net IRC](freenode.net) and in many open channels in
community slacks like [#golang on macadmins.org](macadmins.slack.com) where
friendly people will help get you introduced to the go language.

***
## Contributing
If you're interested in contributing to this project, we would be honored and
happy to have your help. If this is your first project, and you're a little
confused on how to get started, be sure to check out
[How to Contribute to Open Source](https://opensource.guide/how-to-contribute/)
for an overview of good habits, and the Github documentation on
[How to create a Pull Request](https://help.github.com/articles/creating-a-pull-request/)
for the technical bits.

It can be scary at first, but don't worry - you'll do fine.

#### Please submit all pull requests to the azusapacificuniversity/snipe-agent repository in the develop branch!

As you're working on bug-fixes or features, please break them out into their
own feature branches and open the pull request against your feature branch. It
makes it much easier to decipher down the road, as you open multiple pull
requests over time, and makes it much easier for us to approve pull requests
quickly.

Another request is that you do not change the current requirements to running
this program. An example, is that you might create a new function to get data
that is useful to your organization. Our request is that that function isn't
required to run natively or is enabled by default, but rather is available to
users if they configure their version for it.

#### Pull Request Guidelines:

A good commit message should describe what changed and why. Snipe-Agent hopes to
use semantic commit messages to streamline the release process and easily
generate changelogs between versions.

Before a pull request can be merged, it must have a pull request title with a
semantic prefix.

Examples of commit messages with semantic prefixes:

    Fixed #<issue number>:  Fixes get_mac_address() for linux users.
    Added #<issue number>: add get_installed_apps() for a list of installed software.

Please reference the issue or feature request your PR is addressing. Github will
automatically link your PR to the issue, which makes it easier to follow the
bugfix/feature path in the future.

Whenever possible, please provide a clear summary of what your PR does, both
from a technical perspective and from a functionality perspective.
