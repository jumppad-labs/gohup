$lastRunTime = [DateTime]::MinValue 

for ($i=1; $i -le 60; $i++)
{
    $lastRunTime = Get-Date

    # Call your batch file here
    Add-Content -Path .\child.txt -Value $lastRunTime

    Start-Sleep 1
}