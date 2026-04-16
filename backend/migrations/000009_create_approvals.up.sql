-- 部署审批流
CREATE TABLE approvals (
    id SERIAL PRIMARY KEY,
    deployment_id INTEGER NOT NULL REFERENCES deployments(id),
    requester_id INTEGER NOT NULL REFERENCES users(id),
    approver_id INTEGER NOT NULL REFERENCES users(id),
    status VARCHAR(10) NOT NULL DEFAULT 'pending',
    comment TEXT DEFAULT '',
    decided_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_approvals_deployment ON approvals (deployment_id);
CREATE INDEX idx_approvals_requester ON approvals (requester_id);
CREATE INDEX idx_approvals_approver ON approvals (approver_id);
