import 'package:flutter_test/flutter_test.dart';
import 'package:todo/services/item_service.dart';

void main() {
  group('ItemService', () {
    group('constructor', () {
      test('creates service with default values', () {
        final service = ItemService();

        expect(service.host, equals('localhost'));
        expect(service.port, equals(8080));
      });

      test('creates service with custom host', () {
        final service = ItemService(host: 'todo.example.com');

        expect(service.host, equals('todo.example.com'));
        expect(service.port, equals(8080));
      });

      test('creates service with custom port', () {
        final service = ItemService(port: 9090);

        expect(service.host, equals('localhost'));
        expect(service.port, equals(9090));
      });
    });

    group('requireSecureConnection', () {
      test('returns false for localhost', () {
        final service = ItemService(host: 'localhost');
        expect(service.requireSecureConnection(), isFalse);
      });

      test('returns false for 127.0.0.1', () {
        final service = ItemService(host: '127.0.0.1');
        expect(service.requireSecureConnection(), isFalse);
      });

      test('returns false for ::1', () {
        final service = ItemService(host: '::1');
        expect(service.requireSecureConnection(), isFalse);
      });

      test('returns false for empty host', () {
        final service = ItemService(host: '');
        expect(service.requireSecureConnection(), isFalse);
      });

      test('returns true for remote host', () {
        final service = ItemService(host: 'todo.example.com');
        expect(service.requireSecureConnection(), isTrue);
      });
    });
  });

  group('ItemException', () {
    test('stores message', () {
      final exception = ItemException('something failed');
      expect(exception.message, equals('something failed'));
      expect(exception.grpcError, isNull);
    });

    test('toString includes message', () {
      final exception = ItemException('something failed');
      expect(exception.toString(), equals('ItemException: something failed'));
    });
  });
}