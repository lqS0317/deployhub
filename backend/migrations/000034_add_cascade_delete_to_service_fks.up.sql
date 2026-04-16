-- builds.service_id 级联删除
ALTER TABLE builds DROP CONSTRAINT IF EXISTS builds_service_id_fkey;
ALTER TABLE builds ADD CONSTRAINT builds_service_id_fkey FOREIGN KEY (service_id) REFERENCES services(id) ON DELETE CASCADE;

-- deployments.service_id 级联删除
ALTER TABLE deployments DROP CONSTRAINT IF EXISTS deployments_service_id_fkey;
ALTER TABLE deployments ADD CONSTRAINT deployments_service_id_fkey FOREIGN KEY (service_id) REFERENCES services(id) ON DELETE CASCADE;

-- service_members.service_id 级联删除
ALTER TABLE service_members DROP CONSTRAINT IF EXISTS service_members_service_id_fkey;
ALTER TABLE service_members ADD CONSTRAINT service_members_service_id_fkey FOREIGN KEY (service_id) REFERENCES services(id) ON DELETE CASCADE;

-- helm_values.service_id 级联删除
ALTER TABLE helm_values DROP CONSTRAINT IF EXISTS helm_values_service_id_fkey;
ALTER TABLE helm_values ADD CONSTRAINT helm_values_service_id_fkey FOREIGN KEY (service_id) REFERENCES services(id) ON DELETE CASCADE;

-- config_templates.service_id 级联删除
ALTER TABLE config_templates DROP CONSTRAINT IF EXISTS config_templates_service_id_fkey;
ALTER TABLE config_templates ADD CONSTRAINT config_templates_service_id_fkey FOREIGN KEY (service_id) REFERENCES services(id) ON DELETE CASCADE;

-- group_service_permissions.service_id 级联删除
ALTER TABLE group_service_permissions DROP CONSTRAINT IF EXISTS group_service_permissions_service_id_fkey;
ALTER TABLE group_service_permissions ADD CONSTRAINT group_service_permissions_service_id_fkey FOREIGN KEY (service_id) REFERENCES services(id) ON DELETE CASCADE;
