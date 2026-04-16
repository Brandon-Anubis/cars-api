-- 001_create_cars.sql
-- Creates the Cars table and supporting index for the Cars API.
-- Column names match the spanner struct tags on domain.Car (UPPER_CASE).

CREATE TABLE IF NOT EXISTS Cars (
  ID          STRING(36)  NOT NULL,
  MAKE        STRING(100) NOT NULL,
  MODEL       STRING(100) NOT NULL,
  YEAR        INT64       NOT NULL,
  COLOR       STRING(50)  NOT NULL,
  VIN         STRING(17),
  MILEAGE     INT64       NOT NULL,
  PRICE_CENTS INT64       NOT NULL,
  CREATED_AT  TIMESTAMP   NOT NULL,
  UPDATED_AT  TIMESTAMP   NOT NULL,
) PRIMARY KEY (ID);

CREATE INDEX IF NOT EXISTS CarsByMakeModel ON Cars(MAKE, MODEL);
CREATE UNIQUE INDEX IF NOT EXISTS CarsByVIN ON Cars(VIN) WHERE VIN IS NOT NULL;