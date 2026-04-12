-- 002_seed_data.up.sql
-- Insert test user (password hashed with bcrypt, cost 12: password123)
-- You'll need to replace this with actual bcrypt hash
INSERT INTO users (id, name, email, password, created_at) VALUES
('550e8400-e29b-41d4-a716-446655440001', 'Test User', 'test@example.com', '$2a$12$abcdefghijklmnopqrstuvwxyz0123456789ABCDEFGH', NOW())
ON CONFLICT (email) DO NOTHING;

-- Insert test project
INSERT INTO projects (id, name, description, owner_id, created_at) VALUES
('550e8400-e29b-41d4-a716-446655440002', 'Sample Project', 'This is a sample project for testing', '550e8400-e29b-41d4-a716-446655440001', NOW())
ON CONFLICT DO NOTHING;

-- Insert test tasks
INSERT INTO tasks (id, title, description, status, priority, project_id, assignee_id, created_at, updated_at) VALUES
('550e8400-e29b-41d4-a716-446655440003', 'Design homepage', 'Create responsive homepage design', 'todo', 'high', '550e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440001', NOW(), NOW()),
('550e8400-e29b-41d4-a716-446655440004', 'Setup database', 'Configure PostgreSQL', 'in_progress', 'high', '550e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440001', NOW(), NOW()),
('550e8400-e29b-41d4-a716-446655440005', 'Write tests', 'Add integration tests', 'done', 'medium', '550e8400-e29b-41d4-a716-446655440002', NULL, NOW(), NOW())
ON CONFLICT DO NOTHING;
