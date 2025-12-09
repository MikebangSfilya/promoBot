INSERT INTO Users_promo(uid, name, role) VALUES (68761694, 'SadBot', 'admin')
   ON CONFLICT (uid) DO UPDATE SET role = 'admin';
