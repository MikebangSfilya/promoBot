INSERT INTO Users(uid, name, role) VALUES (680806786, 'Sfilya', 'admin')
   ON CONFLICT (uid) DO UPDATE SET role = 'admin';