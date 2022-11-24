--     Put 3 empty lines breaks between each SQL request
--     DO NOT MODIFY / DELETE THIS FILE ONCE COMITTED EXCEPT FOR SQL SYNTAX UPDATES

-- Description: Create base tables for subdomains, certificates, users, toasters and blocked emails

CREATE TABLE IF NOT EXISTS `subdomains` ( 
    `id` VARCHAR(32) NOT NULL ,
    `name` VARCHAR(255) NOT NULL UNIQUE ,
    `user_id` VARCHAR(32) NOT NULL ,
    `toaster_id` VARCHAR(128) NOT NULL ,
    PRIMARY KEY (`id`)) ENGINE = InnoDB;



CREATE TABLE IF NOT EXISTS `certificates` ( 
    `domain` VARCHAR(255) NOT NULL ,
    `cert` BLOB(12000) NOT NULL ,
    PRIMARY KEY (`domain`)) ENGINE = InnoDB;



CREATE TABLE IF NOT EXISTS `users` (
    `cursor` INT NOT NULL AUTO_INCREMENT ,
    `id` VARCHAR(32) NOT NULL ,
    `email` VARCHAR(320) NOT NULL UNIQUE ,
    `username` VARCHAR(320) NOT NULL UNIQUE ,
    `picture_ext` VARCHAR(32) NOT NULL DEFAULT '',
    `password` VARCHAR(128) NOT NULL ,
    PRIMARY KEY (`id`) ,
    KEY (`cursor`)) ENGINE = InnoDB;



CREATE TABLE IF NOT EXISTS `toasters` (
    `cursor` INT NOT NULL AUTO_INCREMENT ,
    `id` VARCHAR(128) NOT NULL ,
    `code_id` VARCHAR(128) NOT NULL ,
    `owner_id` VARCHAR(128) NOT NULL ,
    `image` VARCHAR(128) NOT NULL ,
    `build_command` MEDIUMBLOB ,
    `execution_command` MEDIUMBLOB ,
    `environment_variables` MEDIUMBLOB ,
    `joinable_for_seconds` INT UNSIGNED ,
    `max_concurrent_joiners` INT UNSIGNED ,
    `timeout_seconds` INT UNSIGNED ,
    `name` VARCHAR(120) ,
    `last_modified` INT UNSIGNED ,
    `created` INT UNSIGNED ,
    `git_url` VARCHAR(500) ,
    `git_username` VARCHAR(60) ,
    `git_branch` VARCHAR(60) ,
    `git_access_token` VARCHAR(300) ,
    `git_password` VARCHAR(120) ,
    `files` MEDIUMBLOB ,
    `readme` MEDIUMTEXT ,
    `keywords` TINYBLOB ,
    `picture_ext` VARCHAR(32) NOT NULL DEFAULT '',
    PRIMARY KEY (`id`) ,
    KEY (`cursor`)) ENGINE = InnoDB;



CREATE TABLE IF NOT EXISTS `email_blocklist` ( 
    `id` INT NOT NULL AUTO_INCREMENT ,
    `date` INT NOT NULL ,
    `email` VARCHAR(320) NOT NULL UNIQUE ,
    `data` VARCHAR(320) ,
    PRIMARY KEY (`id`)) ENGINE = InnoDB;



CREATE TABLE IF NOT EXISTS `user_statistics` ( 
    `user_id` VARCHAR(32) NOT NULL ,
    `month_year` VARCHAR(16) NOT NULL ,
    `duration_ms` INT ,
    `cpus` INT ,
    `executions` INT ,
    `ram_gbs` DOUBLE UNSIGNED ,
    `net_ingress` DOUBLE UNSIGNED ,
    `net_egress` DOUBLE UNSIGNED ,
    PRIMARY KEY (`user_id`, `month_year`)) ENGINE = InnoDB;