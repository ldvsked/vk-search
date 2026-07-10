INSERT INTO roles (name)
VALUES
    ('reader'),
    ('editor'),
    ('admin')
ON CONFLICT (name) DO NOTHING;