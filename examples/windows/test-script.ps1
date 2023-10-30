$lastRunTime = [DateTime]::MinValue 

# Start a child script
& "$PSScriptRoot\child-script.ps1"

for ($i=1; $i -le 60; $i++)
{
    $lastRunTime = Get-Date

    # Call your batch file here
    Add-Content -Path .\out.txt -Value $lastRunTime

    Start-Sleep 1
}