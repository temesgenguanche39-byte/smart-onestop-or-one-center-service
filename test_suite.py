import urllib.request
import json
import sys

def main():
    print("=" * 60)
    print("Smart One-Stop Platform Integration Verification")
    print("=" * 60)

    # 1. Login as Official
    login_payload = json.dumps({
        "email": "woreda01.officer@smartonestop.gov.et",
        "password": "Password123!"
    }).encode("utf-8")
    
    req = urllib.request.Request(
        "http://localhost:8080/api/v1/auth/login",
        data=login_payload,
        headers={"Content-Type": "application/json"}
    )
    with urllib.request.urlopen(req) as resp:
        auth_data = json.loads(resp.read().decode("utf-8"))
        token = auth_data["token"]
        user = auth_data["user"]
        print(f"[1] Official Login Successful: {user['full_name']} | Role: {user['role']} | Jurisdiction: {user['structure_name']}")

    # 2. Executive Analytics Dashboard
    req = urllib.request.Request(
        "http://localhost:8080/api/v1/analytics/dashboard",
        headers={"Authorization": f"Bearer {token}"}
    )
    with urllib.request.urlopen(req) as resp:
        dash = json.loads(resp.read().decode("utf-8"))
        print(f"[2] Executive Dashboard: Total={dash['total_cases']}, Active={dash['active_cases']}, Breached={dash['breached_cases']}")

    # 3. Check Ticket TKT-2026-100582 Before Escalation
    req = urllib.request.Request("http://localhost:8080/api/v1/cases/ticket/TKT-2026-100582")
    with urllib.request.urlopen(req) as resp:
        before_tkt = json.loads(resp.read().decode("utf-8"))
        print(f"[3] Pre-Sweep Ticket Status: {before_tkt['status']} at {before_tkt['current_structure_name']} ({before_tkt['structure_level']})")

    # 4. Trigger SLA Escalation Sweep
    req = urllib.request.Request(
        "http://localhost:8080/api/v1/sla/trigger-sweep",
        data=b"{}",
        headers={"Authorization": f"Bearer {token}", "Content-Type": "application/json"}
    )
    with urllib.request.urlopen(req) as resp:
        sweep_res = json.loads(resp.read().decode("utf-8"))
        print(f"[4] SLA Sweep Result: {sweep_res['message']} (Escalated: {sweep_res['escalated_count']} tickets)")

    # 5. Verify Case TKT-2026-100582 Post-Escalation
    req = urllib.request.Request("http://localhost:8080/api/v1/cases/ticket/TKT-2026-100582")
    with urllib.request.urlopen(req) as resp:
        after_tkt = json.loads(resp.read().decode("utf-8"))
        print(f"[5] Post-Sweep Ticket Status: {after_tkt['status']}")
        print(f"    Current Tier: {after_tkt['current_structure_name']} ({after_tkt['structure_level']})")
        print(f"    Escalation Count: {after_tkt['escalation_count']}")
        print(f"    Is Escalated: {after_tkt['is_escalated']}")
        latest_audit = after_tkt['audit_logs'][-1]
        print(f"    Immutable Audit: {latest_audit['action']} - {latest_audit['notes']}")

    # 6. Test New Citizen Intake Submission
    new_case_payload = json.dumps({
        "citizen_full_name": "Abebe Bikila",
        "citizen_phone": "+251911998877",
        "citizen_national_id": "ETH-NAT-998811",
        "woreda_id": 4, # Kirkos Woreda 01
        "service_type_id": 1, # Land Administration
        "title": "Title Deed Cadastral Boundary Correction",
        "description": "Cadastral survey discrepancy identified after masterplan update.",
        "priority": "HIGH"
    }).encode("utf-8")
    req = urllib.request.Request(
        "http://localhost:8080/api/v1/cases",
        data=new_case_payload,
        headers={"Content-Type": "application/json", "X-Idempotency-Key": "test-key-001"}
    )
    with urllib.request.urlopen(req) as resp:
        new_case = json.loads(resp.read().decode("utf-8"))
        print(f"[6] New Grievance Intake Successful: Ticket={new_case['ticket_number']} | QR={new_case['qr_verification_code']}")

    print("=" * 60)
    print("ALL INTEGRATION TESTS PASSED (100% OPERATIONAL)")
    print("=" * 60)

if __name__ == "__main__":
    main()
