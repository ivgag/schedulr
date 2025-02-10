
CREATE TABLE linked_accounts (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    provider VARCHAR(50) NOT NULL,           
    access_token VARCHAR(255),               
    refresh_token VARCHAR(255),              
    token_expires_at TIMESTAMP,              
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (user_id, provider),
    FOREIGN KEY (user_id) REFERENCES users(id)
);