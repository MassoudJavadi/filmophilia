-- Revert primary key constraint names
ALTER INDEX notifications_pkey RENAME TO notifications_partitioned_pkey;
ALTER INDEX activities_pkey RENAME TO activities_partitioned_pkey;
