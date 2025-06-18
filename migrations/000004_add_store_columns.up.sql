-- Enable uuid-ossp extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Add category_id column
ALTER TABLE stores ADD COLUMN IF NOT EXISTS category_id UUID;

-- Add address column
ALTER TABLE stores ADD COLUMN IF NOT EXISTS address VARCHAR(255);

-- Add delivery_radius column
ALTER TABLE stores ADD COLUMN IF NOT EXISTS delivery_radius INTEGER;

-- Update id column to use UUID instead of SERIAL
ALTER TABLE stores ALTER COLUMN id TYPE UUID USING (uuid_generate_v4());
ALTER TABLE stores ALTER COLUMN id SET DEFAULT uuid_generate_v4(); 