import 'package:flutter_test/flutter_test.dart';
import 'package:todo/proto/item.pb.dart';
import 'package:todo/services/intent_service.dart';

void main() {
  group('IntentService.parseDeepLink', () {
    final createTests = <({
      String link,
      String? title,
      String? description,
      List<String> labels,
    })>[
      (link: 'todo://item/new', title: null, description: null, labels: []),
      (link: 'todo://item/new/', title: null, description: null, labels: []),
      (
        link: 'todo://item/new?title=Buy%20milk&description=2L%20%26%20eggs',
        title: 'Buy milk',
        description: '2L & eggs',
        labels: [],
      ),
      (
        link: 'todo://item/new?title=%20%20&labels=home,%20errand,,home',
        title: null,
        description: null,
        labels: ['home', 'errand'],
      ),
      (
        // As built by the add-todo Gallery skill, which encodes the comma
        link: 'todo://item/new?title=Renew%20passport&labels=admin%2Ctravel',
        title: 'Renew passport',
        description: null,
        labels: ['admin', 'travel'],
      ),
    ];
    for (final tc in createTests) {
      test(tc.link, () {
        final intent = IntentService.parseDeepLink(tc.link);
        if (intent is! CreateItemIntent) {
          fail('expected CreateItemIntent but got $intent');
        }
        expect(intent.title, tc.title);
        expect(intent.description, tc.description);
        expect(intent.labels, tc.labels);
      });
    }

    final viewTests = <({String link, ItemView view})>[
      (link: 'todo://items', view: ItemView.ITEM_VIEW_TRIAGED),
      (link: 'todo://items?view=untriaged', view: ItemView.ITEM_VIEW_UNTRIAGED),
      (link: 'todo://items?view=triaged', view: ItemView.ITEM_VIEW_TRIAGED),
      (
        link: 'todo://items?view=Time-Sensitive',
        view: ItemView.ITEM_VIEW_TIME_SENSITIVE,
      ),
      (link: 'todo://items?view=done', view: ItemView.ITEM_VIEW_DONE),
      (link: 'todo://items?view=unknown', view: ItemView.ITEM_VIEW_TRIAGED),
    ];
    for (final tc in viewTests) {
      test(tc.link, () {
        final intent = IntentService.parseDeepLink(tc.link);
        if (intent is! OpenViewIntent) {
          fail('expected OpenViewIntent but got $intent');
        }
        expect(intent.view, tc.view);
      });
    }

    test('todo://item/42', () {
      final intent = IntentService.parseDeepLink('todo://item/42');
      if (intent is! OpenItemIntent) {
        fail('expected OpenItemIntent but got $intent');
      }
      expect(intent.id, 42);
    });

    for (final link in [
      '',
      'not a link',
      'https://item/new',
      'expensetracker://entry/new',
      'todo://item',
      'todo://item/abc',
      'todo://item/0',
      'todo://item/42/comments',
      'todo://items/extra',
      'todo://labels',
    ]) {
      test('rejects "$link"', () {
        expect(IntentService.parseDeepLink(link), isNull);
      });
    }
  });

  group('IntentService.parseSharedText', () {
    final longTitle = List.filled(30, 'word').join(' '); // 149 characters
    final tests = <({
      String name,
      String text,
      String? subject,
      String? title,
      String? description,
    })>[
      (
        name: 'single line',
        text: '  Call the   plumber ',
        subject: null,
        title: 'Call the plumber',
        description: null,
      ),
      (
        name: 'multiple lines',
        text: 'Call plumber\nabout the leak\n\nunder the sink',
        subject: null,
        title: 'Call plumber',
        description: 'about the leak\n\nunder the sink',
      ),
      (
        name: 'subject and URL from a browser',
        text: 'https://go.dev/blog/',
        subject: 'The Go Blog',
        title: 'The Go Blog',
        description: 'https://go.dev/blog/',
      ),
      (
        name: 'blank subject is ignored',
        text: 'Renew passport',
        subject: '  ',
        title: 'Renew passport',
        description: null,
      ),
      (
        name: 'lone URL',
        text: 'https://github.com/alexhokl/todo-cli/issues/1',
        subject: null,
        title: 'github.com',
        description: 'https://github.com/alexhokl/todo-cli/issues/1',
      ),
      (
        name: 'long line is split at a word boundary',
        text: '$longTitle\nmore',
        subject: null,
        title: List.filled(24, 'word').join(' '),
        description: '${List.filled(6, 'word').join(' ')}\n\nmore',
      ),
      (
        name: 'blank text',
        text: ' \n ',
        subject: null,
        title: null,
        description: null,
      ),
    ];
    for (final tc in tests) {
      test(tc.name, () {
        final intent =
            IntentService.parseSharedText(tc.text, subject: tc.subject);
        expect(intent.title, tc.title);
        expect(intent.description, tc.description);
        expect(intent.labels, isEmpty);
      });
    }
  });

  group('IntentService.toAppIntent', () {
    test('send uses text and subject', () {
      final intent = IntentService.toAppIntent(
        {'action': 'send', 'text': 'https://go.dev', 'subject': 'Go'},
      );
      if (intent is! CreateItemIntent) {
        fail('expected CreateItemIntent but got $intent');
      }
      expect(intent.title, 'Go');
      expect(intent.description, 'https://go.dev');
    });

    test('process_text uses text', () {
      final intent = IntentService.toAppIntent(
        {'action': 'process_text', 'text': 'Buy milk'},
      );
      if (intent is! CreateItemIntent) {
        fail('expected CreateItemIntent but got $intent');
      }
      expect(intent.title, 'Buy milk');
    });

    test('view parses the uri', () {
      final intent = IntentService.toAppIntent(
        {'action': 'view', 'uri': 'todo://items?view=done'},
      );
      expect(intent, isA<OpenViewIntent>());
    });

    test('unknown action is ignored', () {
      expect(IntentService.toAppIntent({'action': 'main'}), isNull);
      expect(IntentService.toAppIntent({}), isNull);
    });
  });
}
