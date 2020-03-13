USE litterbox;

INSERT INTO pickupTimeSlots (date, startTime, endTime, capacity, count)
VALUES ("2020-06-08", "12:00:00", "02:00:00", 10, 8),
       ("2020-06-05", "12:00:00", "02:00:00", 8, 8),
       ("2020-06-05", "10:00:00", "12:00:00", 8, 2),
       ("2020-06-13", "10:00:00", "12:00:00", 10, 0),
       ("2020-06-13", "12:00:00", "02:00:00", 10, 7),
       ("2020-06-14", "10:00:00", "12:00:00", 10, 2),
       ("2020-06-14", "12:00:00", "02:00:00", 10, 2),
       ("2020-06-15", "12:00:00", "02:00:00", 10, 1),
       ("2020-06-15", "10:00:00", "12:00:00", 10, 1),
       ("2020-06-16", "12:00:00", "02:00:00", 10, 1),
       ("2020-06-17", "12:00:00", "02:00:00", 10, 8),
       ("2020-06-18", "10:00:00", "12:00:00", 10, 0),
       ("2020-06-18", "12:00:00", "02:00:00", 10, 8);

INSERT INTO storageItems (name, price)
VALUES ("Other", 25),
       ("Large box", 40),
       ("Book box", 25),
       ("Bin", 40),
       ("Fridge", 30),
       ("Bike", 50),
       ("Television", 50);
