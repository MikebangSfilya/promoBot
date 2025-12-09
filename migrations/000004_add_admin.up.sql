INSERT INTO Users(uid, name, role) VALUES (924283576, 'Anya', 'admin')
   ON CONFLICT (uid) DO UPDATE SET role = 'admin';

