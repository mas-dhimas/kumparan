-- Drop the index on articles.author_id
DROP INDEX IF EXISTS idx_articles_author_id;

-- Drop the articles table
DROP TABLE IF EXISTS articles;

-- Drop the authors table
DROP TABLE IF EXISTS authors;

-- Optionally, drop the uuid-ossp extension (only if you want to fully clean up)
DROP EXTENSION IF EXISTS "uuid-ossp";
