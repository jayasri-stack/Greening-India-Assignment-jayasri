-- 002_seed_data.up.sql
-- Seed test data required by assignment

-- Test user
-- Email: test@example.com
-- Password: password123 (bcrypt hash, cost 12)

INSERT INTO users (id, name, email, password, created_at) VALUES
(
  '550e8400-e29b-41d4-a716-446655440001',
  'Test User',
  'test@example.com',
  '$2a$12$C6UzMDM.H6dfI/f/IKcEeO5tq3cGz9hKpP2rjV0t1Ik5dUvNbzH8W',
  NOW()
)
ON CONFLICT (email) DO NOTHING;

-- Sample project
INSERT INTO projects (id, name, description, owner_id, created_at) VALUES
(
  '550e8400-e29b-41d4-a716-446655440002',
  'Sample Project',
  'This is a sample project for testing',
  '550e8400-e29b-41d4-a716-446655440001',
  NOW()
)
ON CONFLICT DO NOTHING;

-- Sample tasks
INSERT INTO tasks (id, title, description, status, priority, project_id, assignee_id, created_at, updated_at) VALUES
(
  '550e8400-e29b-41d4-a716-446655440003',
  'Design homepage',
  'Create responsive homepage design',
  'todo',
  'high',
  '550e8400-e29b-41d4-a716-446655440002',
  '550e8400-e29b-41d4-a716-446655440001',
  NOW(),
  NOW()
),
(
  '550e8400-e29b-41d4-a716-446655440004',
  'Setup database',
  'Configure PostgreSQL',
  'in_progress',
  'high',
  '550e8400-e29b-41d4-a716-446655440002',
  '550e8400-e29b-41d4-a716-446655440001',
  NOW(),
  NOW()
),
(
  '550e8400-e29b-41d4-a716-446655440005',
  'Write tests',
  'Add integration tests',
  'done',
  'medium',
  '550e8400-e29b-41d4-a716-446655440002',
  NULL,
  NOW(),
  NOW()
)
ON CONFLICT DO NOTHING;