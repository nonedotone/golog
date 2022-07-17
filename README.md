# golog
golang log package

## install
```bash
go get -u github.com/nonedotone/golog
```

## example
```golang
Debug("Debug", "log")
Info("Info", "log")
Error("Error", "log")
Warn("Warn", "log")
Debugf("%s-%s\n", "Debugf", "log")
Infof("%s-%s\n", "Infof", "log")
Errorf("%s-%s\n", "Errorf", "log")
Warnf("%s-%s\n", "Warnf", "log")
Fatal("Fatal")
Fatalf("%s-%s\n", "Fatalf", "log")
```