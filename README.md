# urlShortener
Simple url shortener application 

Обработка конфликтов для эндпоинта POST / упрощена: при попытке повторно сократить URL возвращается статус 201 и новый сгенерированный ключ, а не статус 409 с существующим ключом