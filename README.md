# ProjectGoLive
Welcome to my page for GOrecycle (Project Go Live)

This is a prototype for a recycling marketplace written in GO



It uses 4 mysql tables, which follows:

CREATE TABLE UserSecret  (ID VARCHAR(6) NOT NULL PRIMARY KEY, Username VARCHAR(20) NOT NULL, Password VARCHAR(50) NOT NULL, IsAdmin VARCHAR(5) NOT NULL, CommentItem VARCHAR(300));

CREATE TABLE UserInfo    (ID VARCHAR(6) NOT NULL PRIMARY KEY, Username VARCHAR(20) NOT NULL, LastLogin VARCHAR(50), DateJoin VARCHAR(50) NOT NULL, CommentItem VARCHAR(300));

CREATE TABLE ItemListing (ID VARCHAR(6) NOT NULL PRIMARY KEY, Username VARCHAR(20) NOT NULL, Name VARCHAR(20), ImageLink VARCHAR(200), DatePosted VARCHAR(30), CommentItem VARCHAR(300), ConditionItem VARCHAR(100), Cat VARCHAR(50), ContactMeetInfo VARCHAR(100), Completion VARCHAR(5) );

CREATE TABLE CommentItem (ID VARCHAR(6) NOT NULL PRIMARY KEY, Username VARCHAR(20) NOT NULL, ForItem VARCHAR(20) NOT NULL, Date VARCHAR(50) NOT NULL, CommentItem VARCHAR(300));



Word2vec code credits to
https://github.com/danieldk/go2vec/blob/8029f60947ae/go2vec.go
I had issues getting the original to work due to cblas issues, copied the portion on loading word2vec binary files and normalizing vectors
