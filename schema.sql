create table bucket (
    `name` varchar(100) not null primary key
);

create table file (
    bucket varchar(100) not null,
    id varchar(100) not null,
    `hash` varchar(100) not null,
    ext varchar(100) not null,
    original text not null,
    unique (bucket, id),
    foreign key (bucket) references bucket(name) on delete cascade
);
