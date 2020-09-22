$pipe = $null
$pipeReader = $null

try {
    $pipe = new-object System.IO.Pipes.NamedPipeClientStream("pglet_pipe_111.events")
    $pipe.Connect()

    $pipeReader = new-object System.IO.StreamReader($pipe)
    #$pipeWriter = new-object System.IO.StreamWriter($pipe)
    #$pipeWriter.AutoFlush = $true

    #$pipeWriter.WriteLine("Boom!")

    for ($i = 0; $i -lt 10; $i++) {
        $line = $pipeReader.ReadLine()
        if ($line -eq $null) {
            break
        }
        Write-Host $line

        Start-Sleep -s 2
    }

} finally {
    $pipeReader.Dispose()
    $pipe.Dispose()
}


