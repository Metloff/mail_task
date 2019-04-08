**Пример использования**
	
    echo -e 'https://golang.org\n/etc/passwd\nhttps://golang.org\nhttps://golang.org' | go run main.go

**Запуск тестов**
    
    go test -v
    
**Запуск бенчмарков:**
	
    go test -bench . -benchmem