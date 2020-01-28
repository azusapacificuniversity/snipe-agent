$ServiceName = "SnipeAgent"
$PrettyServiceName = "Snipe Update Agent"
$Description = "This agent updates updates asset information in a Snipe-IT instance."
$RegistryPath = "HKLM:\SYSTEM\CurrentControlSet\Services\$ServiceName"
$srvanyPath = "C:\Windows\srvany.exe"
$AgentPath = "C:\Windows\snipe-agent.exe"

# Move The files into place.
Copy-Item .\snipe-agent.exe -Destination $AgentPath -Force
Copy-Item .\srvany.exe -Destination $srvanyPath -Force

# Create the service
sc.exe create $ServiceName binpath= $srvanyPath type= own start= auto DisplayName= $PrettyServiceName
#Set a description for the service.
sc.exe description $ServiceName $Description


# See if the Registry Path exists now that the service is registered. If not, something went wrong, so quit.
if (!(Test-Path $RegistryPath)){
    Write-Verbose -Verbose "Registry Key is missing!"
    exit 1
}

# See if the parameters keys exists... if not create it and the required Application key.
if (!(Test-Path "$RegistryPath\Parameters")){
    New-Item -Path "$RegistryPath\Parameters" | New-ItemProperty -Name 'Application' -PropertyType 'String' -Value $AgentPath | Out-Null
}else{
    Get-Item -Path "$RegistryPath\Parameters" | New-ItemProperty -Name 'Application' -PropertyType 'String' -Value $AgentPath -Force | Out-Null
}

# Start the newly registered service
Start-Service $ServiceName

# Get the version number and write it as a registry key in the service.
$InstalledVersion = & $AgentPath "-version"
Get-Item $RegistryPath | New-ItemProperty -Name 'Version' -PropertyType 'String' -Value $InstalledVersion -Force | Out-Null
