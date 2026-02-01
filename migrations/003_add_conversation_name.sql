-- Migration: Add name column to conversations for group chat support

ALTER TABLE conversations ADD COLUMN name VARCHAR(100);
