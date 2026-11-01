import unittest
from unittest.mock import patch
import advance


class DemoStatusTests(unittest.TestCase):
    def test_disabled_by_default_and_in_production(self):
        for env, file_env in [({}, {}), ({"APP_ENV": "production", "FOODHUB_DEMO_STATUS": "1"}, {}),
                              ({"APP_ENV": "development", "FOODHUB_DEMO_STATUS": "1"}, {"APP_ENV": "production"}),
                              ({"APP_ENV": "development", "FOODHUB_DEMO_STATUS": "1", "GO_ENV": "production"}, {})]:
            with self.subTest(env=env), self.assertRaises(ValueError):
                advance.check_local(env, file_env)
        advance.check_local({"APP_ENV": "development", "FOODHUB_DEMO_STATUS": "1"}, {})

    def test_customer_cannot_send_status_write(self):
        with patch.object(advance, "request", return_value={"role": "customer"}) as request:
            with self.assertRaises(ValueError):
                advance.advance(22, "confirmed", "customer-token")
            self.assertEqual(request.call_count, 1)
            self.assertEqual(request.call_args.args[0], "GET")

    def test_staff_uses_existing_authenticated_endpoint(self):
        for role in ("manager", "admin"):
            with self.subTest(role=role), patch.object(advance, "request", side_effect=[{"role": role}, {"status": "confirmed"}]) as request:
                self.assertEqual(advance.advance(22, "confirmed", "staff-token"), {"status": "confirmed"})
                self.assertEqual(request.call_args.args, ("PATCH", "/orders/22/status", "staff-token", {"status": "confirmed"}))

    def test_no_cancellation_or_invalid_order_ids(self):
        with patch.object(advance, "request") as request:
            for order_id, status in ((0, "confirmed"), (1, "cancelled"), (1, "complete")):
                with self.assertRaises(ValueError):
                    advance.advance(order_id, status, "staff-token")
            request.assert_not_called()

    def test_transport_is_loopback_and_does_not_follow_redirects(self):
        with patch.object(advance.http.client, "HTTPConnection") as constructor:
            connection = constructor.return_value
            response = connection.getresponse.return_value
            response.status = 302
            response.read.return_value = b'{"success": false, "message": "redirect"}'
            with self.assertRaises(ValueError):
                advance.request("GET", "/auth/me", "staff-token")
            constructor.assert_called_once_with("127.0.0.1", 8080, timeout=10)
            self.assertEqual(connection.request.call_count, 1)
            self.assertEqual(connection.request.call_args.args[:2], ("GET", "/api/auth/me"))
            connection.close.assert_called_once()
