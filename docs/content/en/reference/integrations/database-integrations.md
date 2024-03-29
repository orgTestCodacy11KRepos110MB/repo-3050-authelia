---
title: "Database Integrations"
description: "A database integration reference guide"
lead: "This section contains a database integration reference guide for Authelia."
date: 2022-11-19T16:47:09+11:00
draft: false
images: []
menu:
  reference:
    parent: "integrations"
weight: 320
toc: true
---

We generally recommend using [PostgreSQL] for a database. If high availability is not a consideration we also support
[SQLite3].


## PostgreSQL

The only current support criteria for [PostgreSQL] at present is that the version you're using is supported by the
[PostgreSQL] developers. See their [Versioning Policy](https://www.postgresql.org/support/versioning/) for more
information.

We generally perform integration testing against the latest supported version of [PostgreSQL] and that is generally the
recommended version for new installations.

## MySQL

[MySQL] and [MariaDB] are both supported as part of the [MySQL] implementation. This is generally discouraged as
[PostgreSQL] is widely considered as a significantly better database engine. If you choose to go with [MySQL], we
recommend specifically using the [MariaDB] backend.

[MySQL] comes with some rigid support requirements in addition to the standard requirements for us supporting a third
party.

1. Must both support the `InnoDB` engine and this engine must be the default engine.
2. Must support the `utf8mb4` charset.
3. Must support the `utf8mb4_unicode_520_ci` collation.
4. Must support maximum index size of no less than 2048 bytes. The default maximum index size for the InnoDB engine is
   3072 bytes on:
    1. [MySQL] [8.0](https://dev.mysql.com/doc/refman/8.0/en/innodb-limits.html) or later.
    2. [MySQL] [5.7](https://dev.mysql.com/doc/refman/5.7/en/innodb-limits.html) provided
         [innodb_large_prefix](#innodb-large-prefixes) or later.
    3. [MariaDB] [10.3](https://mariadb.com/kb/en/innodb-system-variables/#innodb_large_prefix) or later.
5. Must support ANSI standard time behaviours. See [ANSI standard time behaviours](#ansi-standard-time-behaviours).

We generally perform integration testing against the latest supported version of [MySQL] and [MariaDB], and the latest
supported version of [MariaDB] is generally the recommended version for new installations.

### Specific Notes

#### InnoDB Large Prefixes

This can be configured in the [MySQL] configuration file by setting the `innodb_large_prefix` value to on.
According to the Oracle documentation this is the default behaviour in
[MySQL] [5.7](https://dev.mysql.com/doc/refman/5.7/en/innodb-parameters.html#sysvar_innodb_large_prefix) and it can't be
turned off in [MySQL] [8.0](https://dev.mysql.com/doc/refman/8.0/en/innodb-limits.html) or in [MariaDB] 10.3 and later.

```cnf
[mysqld]
innodb_large_prefix = ON
```

#### ANSI standard time behaviours

This can be configured in the [MySQL] configuration file by setting the `explicit_defaults_for_timestamp` value to on.
According to the Oracle documentation this is the default behaviour in
[MySQL] [5.7](https://dev.mysql.com/doc/refman/5.7/en/server-system-variables.html#sysvar_explicit_defaults_for_timestamp)
and [MySQL] [8.0](https://dev.mysql.com/doc/refman/8.0/en/server-system-variables.html#sysvar_explicit_defaults_for_timestamp).
This is however not the default behaviour in
[MariaDB](https://mariadb.com/kb/en/server-system-variables/#explicit_defaults_for_timestamp) before 10.10.

```cnf
[mysqld]
explicit_defaults_for_timestamp = ON
```

### Vendor Supported Versions

#### MariaDB Vendor Supported Versions

See the [MariaDB Server Releases](https://mariadb.com/kb/en/mariadb-server-release-dates/) for more information.

#### MySQL Vendor Supported Versions

See the [MySQL Supported Platforms](https://www.mysql.com/support/supportedplatforms/database.html) for information on
which versions and platforms they support.

[PostgreSQL]: https://www.postgresql.org/
[MySQL]: https://www.mysql.com/
[MariaDB]: https://mariadb.org/
[SQLite3]: https://www.sqlite.org/index.html

