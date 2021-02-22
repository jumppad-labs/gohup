$lastRunTime = [DateTime]::MinValue 

for ($i=1; $i -le 10; $i++)
{
    $lastRunTime = Get-Date

    # Call your batch file here
    Add-Content -Path .\out.txt -Value $lastRunTime

    Start-Sleep 1
}