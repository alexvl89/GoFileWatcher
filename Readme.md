Используем следующие библиотеки:

1. github.com/kardianos/service для создания Windows сервиса.
1. github.com/fsnotify/fsnotify для отслеживания изменений в файловой системе.

Выполняем команды: 

```
1. go get github.com/kardianos/service
1. go get github.com/fsnotify/fsnotify
```


# Установка сервиса
go build -o GoFileWatcher.exe
GoFileWatcher.exe install

# Запуск сервиса
GoFileWatcher.exe start



{
  "watch_directory": "C:\\Logs",
  "target_directory": "\\\\network\\path\\to\\copy",
  "file_extensions": [".txt", ".jpg", ".png"]
}
