"""
Smart One-Stop Municipal Civic Platform | የ አንድ ማዕከል አገልግሎት
Role-Based Access Control (RBAC) & Administrative Jurisdiction Scoping Model

ይህ ሞጁል ተጠቃሚዎች በስራ ድርሻቸው (Role) እና በተመደቡበት የስልጣን ወሰን (Jurisdiction) 
ብቻ ኬዞችን እንዲያዩ እና እንዲያስተዳድሩ የሚያስችል የPython ዳታ ሞዴል እና የፍቃድ ቁጥጥር ስርዓት ነው።
"""

import enum
import hashlib
import hmac
import os
import secrets
import sys
from typing import Optional, List, Dict, Any

# Ensure UTF-8 output on Windows terminals
if sys.stdout.encoding != 'utf-8':
    try:
        sys.stdout.reconfigure(encoding='utf-8')
    except Exception:
        pass



# =====================================================================
# 1. የስራ ድርሻ (RBAC Roles)
# =====================================================================
class Role(str, enum.Enum):
    WOREDA_OFFICER = "WOREDA_OFFICER"          # የደረጃ 1 ቅበላ ኦፊሰር (Woreda Level)
    SUBCITY_MANAGER = "SUBCITY_MANAGER"        # የክፍለ ከተማ መሪ (Sub-City Level)
    CITY_DIRECTOR = "CITY_DIRECTOR"            # የከተማ አቀፍ ዳይሬክተር (City-Wide Level)
    SERVICE_DESK_AGENT = "SERVICE_DESK_AGENT"  # ኪዮስክ / ዴስክ ኦፊሰር (Front Desk)
    SUPER_ADMIN = "SUPER_ADMIN"                # ሙሉ አስተዳዳሪ (Global Administration)

    def amharic_title(self) -> str:
        titles = {
            Role.WOREDA_OFFICER: "የደረጃ 1 ቅበላ ኦፊሰር (ወረዳ ደረጃ)",
            Role.SUBCITY_MANAGER: "የክፍለ ከተማ መሪ (የይግባኝና ቅሬታ ሰሚ)",
            Role.CITY_DIRECTOR: "የከተማ አቀፍ ዳይሬክተር (የከተማ ውሳኔ ሰጪ)",
            Role.SERVICE_DESK_AGENT: "ኪዮስክ / ዴስክ ኦፊሰር (ቅበላ)",
            Role.SUPER_ADMIN: "ሙሉ ሲስተም አስተዳዳሪ (Super Administrator)",
        }
        return titles.get(self, str(self.value))


# =====================================================================
# 2. የይለፍ ቃል አመሳጠር (Secure PBKDF2 Password Hashing)
# =====================================================================
class PasswordHasher:
    @staticmethod
    def hash_password(password: str) -> str:
        """Secure PBKDF2-HMAC-SHA256 password hashing with salt"""
        salt = secrets.token_hex(16)
        key = hashlib.pbkdf2_hmac(
            hash_name="sha256",
            password=password.encode("utf-8"),
            salt=salt.encode("utf-8"),
            iterations=100_000
        )
        return f"{salt}:{key.hex()}"

    @staticmethod
    def verify_password(stored_hash: str, password: str) -> bool:
        """Verifies password against stored salt:hash string"""
        try:
            salt, original_hex = stored_hash.split(":")
            new_key = hashlib.pbkdf2_hmac(
                hash_name="sha256",
                password=password.encode("utf-8"),
                salt=salt.encode("utf-8"),
                iterations=100_000
            )
            return hmac.compare_digest(original_hex, new_key.hex())
        except Exception:
            return False


# =====================================================================
# 3. የሰራተኛ አካውንት ክፍል (Staff Account with Jurisdiction Validation)
# =====================================================================
class StaffUser:
    def __init__(
        self,
        full_name: str,
        email: str,
        phone: str,
        password: str,
        role: Role,
        sub_city: Optional[str] = None,
        woreda: Optional[str] = None,
    ):
        self.full_name = full_name
        self.email = email
        self.phone = phone
        self.password_hash = PasswordHasher.hash_password(password)
        self.role = role
        self.sub_city = sub_city.strip() if sub_city else None
        self.woreda = woreda.strip() if woreda else None

        # የስልጣን ወሰን ማረጋገጫ (Jurisdiction Validation)
        self._validate_jurisdiction()

    def _validate_jurisdiction(self):
        """
        የስራ ድርሻው ከሚጠይቀው የስልጣን ወሰን (Jurisdiction) ጋር መጣጣሙን ያረጋግጣል።
        - WOREDA_OFFICER: ክፍለ ከተማ እና ወረዳ ሁለቱም ግዴታ ናቸው።
        - SUBCITY_MANAGER: ክፍለ ከተማ ግዴታ ነው፤ ወረዳ ወደ None ይቀየራል (በስሩ ያሉትን ሁሉ ያያል)።
        - SUPER_ADMIN / CITY_DIRECTOR: ከተማ አቀፍ ሙሉ ስልጣን።
        - SERVICE_DESK_AGENT: ቢያንስ ክፍለ ከተማ ወይም ወረዳ መመደብ አለበት።
        """
        if self.role == Role.WOREDA_OFFICER:
            if not self.sub_city or not self.woreda:
                raise ValueError("ለ WOREDA_OFFICER ክፍለ ከተማ እና ወረዳ መመረጥ አለበት! (Sub-City and Woreda are required for WOREDA_OFFICER)")

        elif self.role == Role.SUBCITY_MANAGER:
            if not self.sub_city:
                raise ValueError("ለ SUBCITY_MANAGER ክፍለ ከተማ መመረጥ አለበት! (Sub-City is required for SUBCITY_MANAGER)")
            self.woreda = None  # በክፍለ ከተማው ስር ያሉትን ሁሉንም ወረዳዎች ያያል

        elif self.role == Role.SERVICE_DESK_AGENT:
            if not self.sub_city and not self.woreda:
                raise ValueError("ለ SERVICE_DESK_AGENT የስራ ወሰን (ክፍለ ከተማ ወይም ወረዳ) መመረጥ አለበት!")

    def check_password(self, candidate_password: str) -> bool:
        return PasswordHasher.verify_password(self.password_hash, candidate_password)

    def get_jurisdiction_summary(self) -> str:
        if self.role in (Role.SUPER_ADMIN, Role.CITY_DIRECTOR):
            return "አዲስ አበባ ከተማ አቀፍ (City-Wide Jurisdiction)"
        if self.role == Role.SUBCITY_MANAGER:
            return f"{self.sub_city} ክፍለ ከተማ (ሁሉንም ወረዳዎች ያጠቃልላል)"
        if self.role == Role.WOREDA_OFFICER:
            return f"{self.sub_city} ክፍለ ከተማ • {self.woreda}"
        if self.role == Role.SERVICE_DESK_AGENT:
            if self.woreda:
                return f"{self.sub_city} ክፍለ ከተማ • {self.woreda} (Service Desk)"
            return f"{self.sub_city} ክፍለ ከተማ (Sub-City Desk)"
        return "ያልተወሰነ የስራ ወሰን"

    def __repr__(self) -> str:
        return (
            f"<StaffUser(name='{self.full_name}', role='{self.role.value}', "
            f"jurisdiction='{self.get_jurisdiction_summary()}')>"
        )


# =====================================================================
# 4. የማዘጋጃ ቤት ኬዝ / ቅሬታ (Municipal Case Model)
# =====================================================================
class MunicipalCase:
    def __init__(self, case_id: str, title: str, sub_city: str, woreda: str, citizen_name: str = "ዜጋ", status: str = "SUBMITTED"):
        self.case_id = case_id
        self.title = title
        self.sub_city = sub_city.strip()
        self.woreda = woreda.strip()
        self.citizen_name = citizen_name
        self.status = status

    def __repr__(self) -> str:
        return f"<Case {self.case_id} | {self.title[:30]}... | 📍 {self.sub_city} • {self.woreda} [{self.status}]>"


# =====================================================================
# 5. የኬዝ ማጣሪያ ስርዓት (Scoping: Only show own authorized cases)
# =====================================================================
class MunicipalService:
    @staticmethod
    def get_accessible_cases(user: StaffUser, all_cases: List[MunicipalCase]) -> List[MunicipalCase]:
        """
        ተጠቃሚው ማየት የሚፈቀድለትን ኬዞች ብቻ ለይቶ ያወጣል (Strict RBAC & Jurisdictional Scoping):
        1. SUPER_ADMIN እና CITY_DIRECTOR: የሁሉንም ክፍለ ከተሞችና ወረዳዎች ያያሉ።
        2. SUBCITY_MANAGER: በራሱ ክፍለ ከተማ ስር ያሉትን ሁሉንም ወረዳዎች ያያል።
        3. WOREDA_OFFICER: በራሱ ክፍለ ከተማ እና በራሱ ወረዳ ያሉትን ብቻ ያያል።
        4. SERVICE_DESK_AGENT: እንደየተመደበበት የስራ ወሰን (ወረዳ ካለው በወረዳው፤ ክፍለ ከተማ ብቻ ከሆነ በክፍለ ከተማው)።
        """
        accessible: List[MunicipalCase] = []

        for case in all_cases:
            # 1. SUPER_ADMIN እና CITY_DIRECTOR የሁሉንም ክፍለ ከተሞችና ወረዳዎች ያያሉ
            if user.role in (Role.SUPER_ADMIN, Role.CITY_DIRECTOR):
                accessible.append(case)

            # 2. SUBCITY_MANAGER በራሱ ክፍለ ከተማ ስር ያሉትን ሁሉንም ወረዳዎች ያያል
            elif user.role == Role.SUBCITY_MANAGER:
                if case.sub_city.lower() == user.sub_city.lower():
                    accessible.append(case)

            # 3. WOREDA_OFFICER በራሱ ክፍለ ከተማ እና በራሱ ወረዳ ያሉትን ብቻ ያያል
            elif user.role == Role.WOREDA_OFFICER:
                if (
                    user.sub_city and user.woreda
                    and case.sub_city.lower() == user.sub_city.lower()
                    and case.woreda.lower() == user.woreda.lower()
                ):
                    accessible.append(case)

            # 4. SERVICE_DESK_AGENT እንደየተመደበበት የስራ ወሰን
            elif user.role == Role.SERVICE_DESK_AGENT:
                if user.woreda and user.sub_city:
                    if case.sub_city.lower() == user.sub_city.lower() and case.woreda.lower() == user.woreda.lower():
                        accessible.append(case)
                elif not user.woreda and user.sub_city:
                    if case.sub_city.lower() == user.sub_city.lower():
                        accessible.append(case)

        return accessible


# =====================================================================
# 6. ማረጋገጫ እና የሙከራ ማስኬጃ (Verification & Simulation Suite)
# =====================================================================
def run_simulation():
    print("=" * 78)
    print(" Smart One-Stop Municipal Platform | RBAC & Jurisdiction Scoping Engine")
    print("=" * 78)

    # 1. የሙከራ ኬዞች (Sample Municipal Cases across Sub-Cities and Woredas)
    cases = [
        MunicipalCase("TKT-2026-001", "የይዞታ ማረጋገጫ ካርታ ጥያቄ", "ቂርቆስ", "ወረዳ 01", "ተስፋዬ በቀለ"),
        MunicipalCase("TKT-2026-002", "የንግድ ፈቃድ እድሳት መጓተት", "ቂርቆስ", "ወረዳ 01", "አበበ ከበደ"),
        MunicipalCase("TKT-2026-003", "የቀበሌ ቤት የኪራይ ውዝግብ", "ቂርቆስ", "ወረዳ 02", "ፋጡማ ሁሴን"),
        MunicipalCase("TKT-2026-004", "የመንገድ መሰረተ ልማት ጥገና", "ቦሌ", "ወረዳ 03", "ዳዊት ታደሰ"),
        MunicipalCase("TKT-2026-005", "የውሃ መስመር መቆራረጥ ቅሬታ", "የካ", "ወረዳ 05", "ሄለን ግርማ"),
    ]

    print(f"\n📦 ጠቅላላ የተመዘገቡ ኬዞች ብዛት: {len(cases)}")
    for c in cases:
        print(f"  • {c}")

    # 2. የሰራተኛ አካውንቶች (Staff Users with Assigned Roles & Jurisdictions)
    super_admin = StaffUser("አለማየሁ ታደሰ", "admin@smartonestop.gov.et", "+251911000001", "Pass123!", Role.SUPER_ADMIN)
    city_director = StaffUser("ዶ/ር ሰላማዊት በቀለ", "director@smartonestop.gov.et", "+251911000002", "Pass123!", Role.CITY_DIRECTOR)
    kirkos_manager = StaffUser("ካሳሁን ኃይሌ", "kirkos.mgr@smartonestop.gov.et", "+251911000003", "Pass123!", Role.SUBCITY_MANAGER, sub_city="ቂርቆስ")
    woreda01_officer = StaffUser("ቤተልሔም ግርማ", "woreda01.off@smartonestop.gov.et", "+251911000004", "Pass123!", Role.WOREDA_OFFICER, sub_city="ቂርቆስ", woreda="ወረዳ 01")
    woreda02_officer = StaffUser("ሙሉቀን አስፋው", "woreda02.off@smartonestop.gov.et", "+251911000005", "Pass123!", Role.WOREDA_OFFICER, sub_city="ቂርቆስ", woreda="ወረዳ 02")
    desk_agent_w01 = StaffUser("ዮናስ ሙሉጌታ", "desk.w01@smartonestop.gov.et", "+251911000006", "Pass123!", Role.SERVICE_DESK_AGENT, sub_city="ቂርቆስ", woreda="ወረዳ 01")

    staff_roster = [super_admin, city_director, kirkos_manager, woreda01_officer, woreda02_officer, desk_agent_w01]

    # 3. የፍቃድ እና የስልጣን ወሰን ማጣሪያ ፍተሻ (Jurisdiction Scoping Verification)
    print("\n" + "-" * 78)
    print("🔍 የፍቃድ ቁጥጥር እና የኬዝ ማጣሪያ ውጤት (Jurisdiction Scoping Matrix)")
    print("-" * 78)

    for user in staff_roster:
        accessible = MunicipalService.get_accessible_cases(user, cases)
        print(f"\n👤 ሰራተኛ: {user.full_name} | {user.role.value}")
        print(f"   📍 የስልጣን ወሰን: {user.get_jurisdiction_summary()}")
        print(f"   👁️ ማየት የሚችሏቸው ኬዞች ({len(accessible)}/{len(cases)}):")
        for ac in accessible:
            print(f"      ✔ [{ac.case_id}] {ac.title} ({ac.sub_city} • {ac.woreda})")

    # 4. የስልጣን ወሰን ህግጋት ማረጋገጫ (Testing Exception Handling on Illegal Jurisdiction)
    print("\n" + "-" * 78)
    print("🛡️ የስልጣን ወሰን ግዴታ ማረጋገጫ (Enforcing Required Sub-City & Woreda)")
    print("-" * 78)

    # ሙከራ 1: WOREDA_OFFICER ያለ ወረዳ ከተፈጠረ ስህተት መጣል አለበት
    try:
        StaffUser("ህገወጥ ኦፊሰር", "invalid@test.com", "+251900000000", "Pass123!", Role.WOREDA_OFFICER, sub_city="ቂርቆስ", woreda=None)
        print("❌ ስህተት: WOREDA_OFFICER ያለ ወረዳ ተፈጠረ!")
    except ValueError as e:
        print(f"✅ ተረጋግጧል [WOREDA_OFFICER]: {e}")

    # ሙከራ 2: SUBCITY_MANAGER ያለ ክፍለ ከተማ ከተፈጠረ ስህተት መጣል አለበት
    try:
        StaffUser("ህገወጥ መሪ", "invalid2@test.com", "+251900000000", "Pass123!", Role.SUBCITY_MANAGER, sub_city=None)
        print("❌ ስህተት: SUBCITY_MANAGER ያለ ክፍለ ከተማ ተፈጠረ!")
    except ValueError as e:
        print(f"✅ ተረጋግጧል [SUBCITY_MANAGER]: {e}")

    print("\n" + "=" * 78)
    print("🎉 የ RBAC እና የ Jurisdiction Scoping ሞዴል በተሟላ መልኩ ሰርቷል!")
    print("=" * 78)


if __name__ == "__main__":
    run_simulation()
