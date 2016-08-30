# SmugMugMugger
Rescue as much as possible of your Picturelife life from your SmugMug account  
After Picturelife shut down, SmugMug holds the files that were on the Picturelife servers.    
This tool might help you extract those files from SmugMug.

The former [rescuelife](https://github.com/morphar/rescuelife) tool won't work anymore.

## Install & run
The binary keeps a file with status of fetches on the disk.  
Because of this, you can restart the application as many times as you like.  

It may cause some downloads to fail, but you should probably retry all failed downloads anyways, after you have successfully run to the end once.

You can safely re-run the application.

### If you are on OS X / macOS
You can download a pre-build binary [here](https://github.com/morphar/SmugMugMugger/releases).  
Or [direct link](https://github.com/morphar/SmugMugMugger/releases/download/0.1.0/SmugMugMugger) to the binary.

In the Finder, double click on the downloaded file.

##### If that doesn't work:
Open Terminal app, change dir to where you downloaded the binary, then run:  
```
cd Downloads        # Takes you to your download dir
chmod +x SmugMugMugger # This will make the file executable
./SmugMugMugger -help  # Help text about what flags can be used
./SmugMugMugger        # This will run the program
```

You can always run the program again to see if any more files has become avialable:
```./SmugMugMugger -retry```

### If you have Go(lang) installed
Get your own SmugMug API key here:
[Apply for an API Key](https://api.smugmug.com/api/developer/apply)


Install:  
```go get github.com/morphar/SmugMugMugger```  

Edit `keys.go` with your own API key and secret.  

Enter the repo:  
```cd $GOPATH/src/github.com/morphar/SmugMugMugger/```

Run:  
```SmugMugMugger -help```

## Notes
In my case, I got more files than the tool reported.  
It get's the number from SmugMug's API though, so I haven't figured out why yet.
