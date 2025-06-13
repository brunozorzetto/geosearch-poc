-- Add delivery_radius column to stores table
ALTER TABLE stores ADD COLUMN delivery_radius INTEGER NOT NULL DEFAULT 5000;

-- Add comment to explain the column
COMMENT ON COLUMN stores.delivery_radius IS 'Maximum delivery radius in meters'; 