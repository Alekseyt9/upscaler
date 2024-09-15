-- авторизация пользователя
CREATE TABLE users (
    id SERIAL PRIMARY KEY,               	-- уникальный идентификатор пользователя
    email VARCHAR(255) UNIQUE,  			-- электронная почта (логин)
    password_hash TEXT,         			-- хеш пароля (для хранения зашифрованного пароля)
    created_at TIMESTAMPTZ DEFAULT NOW(),	-- дата создания учетной записи
    last_login TIMESTAMPTZ,              	-- время последнего входа
    token_expiration TIMESTAMPTZ         	-- время истечения срока действия JWT токена
);

-- глобальная очередь задач
CREATE TABLE queue (
    id SERIAL PRIMARY KEY,   			    -- уникальный идентификатор элемента очереди
    order_num SERIAL              			-- порядок (автоинкремент)
);

-- файлы пользователя
CREATE TABLE userfiles (
    id SERIAL PRIMARY KEY,            		-- уникальный идентификатор файла
    queue_id INTEGER REFERENCES queue(id),  -- ссылка на очередь (может быть NULL)
    user_id INTEGER NOT NULL,          		-- идентификатор пользователя
    order_num SERIAL,                  		-- порядок
    src_file_url TEXT NOT NULL,             -- URL исходного файла
    src_file_key TEXT NOT NULL,             -- ключ исходного файла
    dest_file_url TEXT NOT NULL,            -- URL обработанного файла
    dest_file_key TEXT NOT NULL,            -- ключ обработанного файла
    state TEXT NOT NULL,				    -- состояние задачи: PENDING/PROCESSED/ERROR/OUTDATED
    created_at TIMESTAMPTZ DEFAULT NOW(),   -- время создания записи

    CONSTRAINT fk_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- transaction outbox для отправки в очередь сообщений
CREATE TABLE outbox (
    id SERIAL PRIMARY KEY,                  -- уникальный идентификатор сообщения
    payload JSONB NOT NULL,                 -- данные для передачи (данные для Kafka)
    status VARCHAR(50) DEFAULT 'PENDING',   -- статус отправки (PENDING, SENT, FAILED)
    created_at TIMESTAMPTZ DEFAULT NOW(),   -- время создания записи
    processed_at TIMESTAMPTZ,               -- время обработки (когда сообщение было отправлено)
    retry_count INTEGER DEFAULT 0,          -- количество попыток отправки
    last_error TEXT                         -- описание последней ошибки (если статус FAILED)
);
