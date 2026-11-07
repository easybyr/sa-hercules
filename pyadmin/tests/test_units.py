"""工具层与业务常量单元测试。"""

from __future__ import annotations

from datetime import datetime, timedelta, timezone

import jwt
import pytest

from app.config import load_settings
from app.consts.patterns import EMAIL_PATTERN, PERMISSION_CODE_PATTERN
from app.consts.status import Status
from app.pkg import apperror, idgen, password, response
from app.pkg.timeutil import to_iso
from app.pkg.token import Claims, TokenManager
from app.pkg.validation import format_validation_errors
from app.schema.pagination import PageRequest, PageResult


class TestIdGenerator:
    def test_numeric_has_expected_digits_and_no_leading_zero(self):
        for digits in (8, 10):
            for _ in range(50):
                value = idgen.numeric(digits)
                assert len(str(value)) == digits
                assert str(value)[0] != "0"

    def test_numeric_rejects_invalid_digits(self):
        with pytest.raises(ValueError):
            idgen.numeric(0)
        with pytest.raises(ValueError):
            idgen.numeric(19)

    def test_random_password_length_and_alphabet(self):
        value = idgen.random_password(16)
        assert len(value) == 16
        assert all(character.isalnum() or character in "!@#$%" for character in value)
        with pytest.raises(ValueError):
            idgen.random_password(8 - 1)

    def test_personal_role_id_is_negative_uid(self):
        assert idgen.personal_role_id(1234567890) == -1234567890


class TestPassword:
    def test_hash_and_verify(self):
        hashed = password.hash_password("admin")
        assert hashed != "admin"
        assert password.verify_password(hashed, "admin")
        assert not password.verify_password(hashed, "wrong")

    def test_verify_handles_empty_and_invalid_hash(self):
        assert not password.verify_password("", "admin")
        assert not password.verify_password("not-a-bcrypt-hash", "admin")

    def test_hash_rejects_oversized_password(self):
        with pytest.raises(ValueError):
            password.hash_password("a" * 73)


class TestToken:
    def test_issue_and_parse_roundtrip(self):
        manager = TokenManager("secret", 12)
        token = manager.issue(1_000_000_000, 10_000_000, "admin", True)
        claims = manager.parse(token)
        assert claims == Claims(uid=1_000_000_000, org_id=10_000_000, is_super=True, username="admin")
        assert manager.expires_in == 12 * 3600

    def test_parse_rejects_wrong_secret(self):
        token = TokenManager("secret-a", 1).issue(1, 1, "user", False)
        with pytest.raises(apperror.AppError) as error:
            TokenManager("secret-b", 1).parse(token)
        assert error.value.status == 401
        assert error.value.message == "登录凭证无效或已过期"

    def test_parse_rejects_expired_token(self):
        manager = TokenManager("secret", 1)
        expired = jwt.encode(
            {
                "uid": 1,
                "org_id": 1,
                "is_super": False,
                "username": "user",
                "sub": "user",
                "iat": int((datetime.now(tz=timezone.utc) - timedelta(hours=3)).timestamp()),
                "exp": int((datetime.now(tz=timezone.utc) - timedelta(hours=2)).timestamp()),
            },
            "secret",
            algorithm="HS256",
        )
        with pytest.raises(apperror.AppError):
            manager.parse(expired)

    def test_parse_rejects_missing_claims(self):
        manager = TokenManager("secret", 1)
        token = jwt.encode(
            {
                "sub": "user",
                "iat": int(datetime.now(tz=timezone.utc).timestamp()),
                "exp": int((datetime.now(tz=timezone.utc) + timedelta(hours=1)).timestamp()),
            },
            "secret",
            algorithm="HS256",
        )
        with pytest.raises(apperror.AppError):
            manager.parse(token)


class TestStatus:
    def test_labels_and_validity(self):
        assert Status.NORMAL.label() == "normal"
        assert Status.DISABLED.label() == "disabled"
        assert Status.label_of(1) == "normal"
        assert Status.label_of(2) == "disabled"
        assert Status.label_of(99) == "unknown"
        assert Status.is_valid(1) and Status.is_valid(2)
        assert not Status.is_valid(0) and not Status.is_valid(3)


class TestPagination:
    def test_defaults_and_clamping(self):
        request = PageRequest()
        assert (request.page, request.page_size, request.offset) == (1, 10, 0)

        clamped = PageRequest(page=0, page_size=5000)
        assert (clamped.page, clamped.page_size) == (1, 1000)

        second = PageRequest(page=3, page_size=20)
        assert second.offset == 40

    def test_page_result_serialization(self):
        result = PageResult[int](list=[1, 2], total=2, page=1, page_size=10)
        assert result.model_dump() == {"list": [1, 2], "total": 2, "page": 1, "page_size": 10}


class TestResponse:
    def test_envelope_contains_three_fields(self):
        assert response.envelope(0, "success", {"a": 1}) == {
            "code": 0,
            "message": "success",
            "data": {"a": 1},
        }

    def test_created_uses_201(self):
        import json

        created = response.created({"id": 1})
        assert created.status_code == 201
        assert json.loads(created.body) == {
            "code": 0,
            "message": "created",
            "data": {"id": 1},
        }

    def test_failure_code_equals_status(self):
        import json

        failed = response.failure(403, "没有权限")
        assert failed.status_code == 403
        assert json.loads(failed.body) == {
            "code": 403,
            "message": "没有权限",
            "data": None,
        }


class TestTimeUtil:
    def test_serializes_with_millisecond_precision(self):
        zone = timezone(timedelta(hours=8))
        moment = datetime(2026, 11, 6, 23, 17, 52, 123456, tzinfo=zone)
        assert to_iso(moment) == "2026-11-06T23:17:52.123+08:00"
        assert to_iso(None) is None

    def test_naive_datetime_is_treated_as_utc(self):
        assert to_iso(datetime(2026, 1, 1, 0, 0, 0)) == "2026-01-01T00:00:00.000+00:00"


class TestPatterns:
    def test_permission_code_pattern(self):
        assert PERMISSION_CODE_PATTERN.match("report.read")
        assert PERMISSION_CODE_PATTERN.match("user.assign_role")
        assert not PERMISSION_CODE_PATTERN.match("Report.read")
        assert not PERMISSION_CODE_PATTERN.match("report")
        assert not PERMISSION_CODE_PATTERN.match("report.")

    def test_email_pattern(self):
        assert EMAIL_PATTERN.match("user@example.com")
        assert not EMAIL_PATTERN.match("user@example")
        assert not EMAIL_PATTERN.match("not-an-email")


class TestValidationMessages:
    def test_missing_field_message(self):
        errors = [{"type": "missing", "loc": ("body", "username"), "msg": "Field required"}]
        assert format_validation_errors(errors) == "用户名不能为空"

    def test_value_error_message_is_unwrapped(self):
        errors = [
            {
                "type": "value_error",
                "loc": ("body", "password"),
                "msg": "Value error, 密码长度需在5到72个字符之间",
            }
        ]
        assert format_validation_errors(errors) == "密码长度需在5到72个字符之间"

    def test_parsing_error_message(self):
        errors = [{"type": "int_parsing", "loc": ("path", "uid"), "msg": "Input should be a valid integer"}]
        assert format_validation_errors(errors) == "用户ID格式不正确"

    def test_empty_errors_fallback(self):
        assert format_validation_errors([]) == "参数校验失败"


class TestConfigLoading:
    def test_dev_config_is_default(self, monkeypatch):
        monkeypatch.delenv("APP_ENV", raising=False)
        monkeypatch.delenv("DATABASE_DSN", raising=False)
        monkeypatch.delenv("JWT_SECRET", raising=False)
        settings = load_settings()
        assert settings.environment == "dev"
        assert settings.config_file.name == "config.dev.yaml"
        assert settings.database.dsn.startswith("postgresql+asyncpg://")

    def test_environment_variables_override_file_values(self, monkeypatch):
        monkeypatch.setenv("APP_ENV", "prod")
        monkeypatch.setenv("DATABASE_DSN", "postgres://user:pass@db:5432/other")
        monkeypatch.setenv("JWT_SECRET", "environment-secret")
        settings = load_settings()
        assert settings.environment == "prod"
        assert settings.database.dsn == "postgresql+asyncpg://user:pass@db:5432/other"
        assert settings.security.jwt_secret == "environment-secret"

    def test_unsupported_environment_raises(self):
        with pytest.raises(ValueError):
            load_settings("staging")
