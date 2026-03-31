# Migration Guide for golang-migrate

# golang-migrate (what this project uses)

migrate create -ext sql -dir database/migrations -seq iam\_\_create_users

# Output:

# 000001_iam\_\_create_users.up.sql

# 000001_iam\_\_create_users.down.sql

# Or with timestamps instead of sequential:

migrate create -ext sql -dir database/migrations iam\_\_create_users

# Output:

# 20260325143022_iam\_\_create_users.up.sql

# 20260325143022_iam\_\_create_users.down.sql
