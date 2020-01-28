$ServiceName = "SnipeAgent"
$srvanyPath = "C:\Windows\srvany.exe"
$AgentPath = "C:\Windows\snipe-agent.exe"

# Stop the Service
Stop-Service $ServiceName

# Unregister and delete the Service
sc.exe delete $ServiceName

# Remove the files.
Remove-Item $srvanyPath -Force
Remove-Item $AgentPath -Force
