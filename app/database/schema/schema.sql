CREATE TABLE AppUser (
    user_id serial,
    first_name text not null,
    last_name text not null,
    email text not null,
    username text not null unique,
    password text not null,
    is_admin bool default false not null,
    creation_time timestamptz NOT NULL default now(),
    last_updated timestamptz,
    workflow_status integer not null default 0,
    max_item_count integer NOT NULL,
    max_item_space float(3) NOT NULL,
    PRIMARY KEY (user_id)
);

CREATE TABLE Groups (
    group_id serial primary key,
    group_name text not null unique,
    created_by integer not null,
    creation_time timestamptz NOT NULL default now(),
    last_updated timestamptz,
    workflow_status integer not null default 0,
    max_item_count integer NOT NULL,
    max_item_space float(3) NOT NULL,
    FOREIGN KEY (created_by) references AppUser(user_id)
);

CREATE TABLE UserGroupMap (
    user_id integer not null ,
    group_id integer not null,
    created_by integer not null,
    is_leader boolean not null default false,
    creation_time timestamptz NOT NULL default now(),
    last_updated timestamptz,
    workflow_status integer not null default 0,
    FOREIGN KEY (user_id) references AppUser(user_id),
    FOREIGN KEY (group_id) references Groups(group_id),
    PRIMARY KEY (user_id, group_id)
);

CREATE TABLE ItemTypes (
    item_type_id integer,
    item_type_name text NOT NULL,
    max_item_count integer NOT NULL,
    max_item_space float(3) NOT NULL,
    PRIMARY KEY (item_type_id)
);

-- Add item types
insert into ItemTypes (item_type_id, item_type_name, max_item_count, max_item_space) values(1, 'Picture', 1, 2), (2, 'Video', 1, 20);

CREATE TABLE Items (
    item_id serial,
    item_name text NOT NULL,
    description text,
    item_type_id integer NOT NULL,
    group_id integer NOT NULL,
    item_path text NOT NULL,
    item_size integer NOT NULL,
    uploaded boolean default false,
    created_by integer NOT NULL,
    creation_time timestamptz NOT NULL default now(),
    last_accessed timestamptz,
    FOREIGN KEY (created_by) references AppUser(user_id),
    FOREIGN KEY (group_id) references Groups(group_id),
    FOREIGN KEY (item_type_id) references ItemTypes(item_type_id),
    PRIMARY KEY (item_id)
);

CREATE TABLE Comments (
    comment_id serial,
    comment text not null,
    item_id integer not null,
    created_by integer not null,
    creation_time timestamptz NOT NULL default now(),
    FOREIGN KEY (created_by) references AppUser(user_id),
    FOREIGN KEY (item_id) references Items(item_id),
    PRIMARY KEY (comment_id)
);

CREATE INDEX idx_Comments_ItemID  ON Comments(item_id);