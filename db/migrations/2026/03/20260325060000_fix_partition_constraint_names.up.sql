-- Fix primary key constraint names for partitioned tables
ALTER INDEX notifications_partitioned_pkey RENAME TO notifications_pkey;
ALTER INDEX activities_partitioned_pkey RENAME TO activities_pkey;
