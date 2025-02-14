
CREATE TABLE linked_accounts (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    provider VARCHAR(50) NOT NULL,           
    access_token VARCHAR(255),               
    refresh_token VARCHAR(255),              
    expiry TIMESTAMP,              
    created_at TIMESTAMP DEFAULT (timezone('utc', now())),
    updated_at TIMESTAMP DEFAULT (timezone('utc', now())),
    UNIQUE (user_id, provider),
    FOREIGN KEY (user_id) REFERENCES users(id)
);