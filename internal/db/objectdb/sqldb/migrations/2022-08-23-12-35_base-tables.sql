--     Put 3 empty lines breaks between each SQL request
--     DO NOT MODIFY / DELETE THIS FILE ONCE COMITTED EXCEPT FOR SQL SYNTAX UPDATES

-- Description: Noop for DB testing purposes, if this one fails it does no damage to the db structure

CREATE TABLE `subdomains` ( 
    `id` VARCHAR(32) NOT NULL ,
    `name` VARCHAR(255) NOT NULL UNIQUE ,
    `user_id` VARCHAR(32) NOT NULL ,
    `toaster_id` VARCHAR(128) NOT NULL ,
    PRIMARY KEY (`id`)) ENGINE = InnoDB;



CREATE TABLE `certificates` ( 
    `domain` VARCHAR(255) NOT NULL ,
    `cert` BLOB(12000) NOT NULL
    PRIMARY KEY (`domain`)) ENGINE = InnoDB;



CREATE TABLE `users` (
    `id` VARCHAR(32) NOT NULL ,
    `email` VARCHAR(320) NOT NULL UNIQUE ,
    `password` VARCHAR(128) NOT NULL ,
    PRIMARY KEY (`id`)
) ENGINE = InnoDB;



CREATE TABLE `toasters` (
    `id` VARCHAR(128) NOT NULL ,
    `code_id` VARCHAR(128) NOT NULL ,
    `owner_id` VARCHAR(128) NOT NULL ,
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
    PRIMARY KEY (`id`)
) ENGINE = InnoDB;



CREATE TABLE `email_blocklist` ( 
    `id` INT NOT NULL AUTO_INCREMENT ,
    `date` INT NOT NULL ,
    `email` VARCHAR(320) NOT NULL UNIQUE ,
    `data` VARCHAR(320) ,
    PRIMARY KEY (`id`)
) ENGINE = InnoDB;