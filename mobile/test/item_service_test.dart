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

  group('normaliseColour', () {
    test('returns null for null input', () {
      expect(ItemService.normaliseColour(null), isNull);
    });

    test('returns null for empty string', () {
      expect(ItemService.normaliseColour(''), isNull);
    });

    test('returns canonical upper-case for valid upper-case input', () {
      expect(ItemService.normaliseColour('#FF0000'), equals('#FF0000'));
    });

    test('canonicalises lower-case hex to upper-case', () {
      expect(ItemService.normaliseColour('#00ff00'), equals('#00FF00'));
    });

    test('canonicalises mixed-case hex to upper-case', () {
      expect(ItemService.normaliseColour('#aBcDeF'), equals('#ABCDEF'));
    });

    test('throws when hash prefix is missing', () {
      expect(() => ItemService.normaliseColour('FF0000'), throwsA(isA<ItemException>()));
    });

    test('throws when too short', () {
      expect(() => ItemService.normaliseColour('#FF0'), throwsA(isA<ItemException>()));
    });

    test('throws when too long', () {
      expect(() => ItemService.normaliseColour('#FF00000'), throwsA(isA<ItemException>()));
    });

    test('throws on non-hex characters', () {
      expect(() => ItemService.normaliseColour('#GGGGGG'), throwsA(isA<ItemException>()));
    });
  });
}