import 'package:flutter_helper/intents.dart';
import 'package:todo/proto/item.pb.dart';

/// An action requested by another app through an Android intent
/// (share sheet, text selection menu, deep link or launcher shortcut).
sealed class AppIntent {
  const AppIntent();
}

/// Opens [EditItemPage] in create mode with the given values pre-filled.
class CreateItemIntent extends AppIntent {
  final String? title;
  final String? description;
  final List<String> labels;

  const CreateItemIntent({
    this.title,
    this.description,
    this.labels = const [],
  });
}

/// Switches the home list to [view].
class OpenViewIntent extends AppIntent {
  final ItemView view;

  const OpenViewIntent(this.view);
}

/// Opens [ItemDetailPage] for the item with [id].
class OpenItemIntent extends AppIntent {
  final int id;

  const OpenItemIntent(this.id);
}

/// Receives intents forwarded by MainActivity over a method channel.
class IntentService {
  static const String deepLinkScheme = 'todo';
  static const String channelName = 'com.alexhokl.todo/intents';

  final IntentChannel<AppIntent> _channel = IntentChannel(
    channelName,
    toAppIntent,
  );

  /// Titles longer than this are cut at a word boundary; the rest of the
  /// line moves to the description.
  static const int maxTitleLength = 120;

  static const Map<String, ItemView> _views = {
    'untriaged': ItemView.ITEM_VIEW_UNTRIAGED,
    'triaged': ItemView.ITEM_VIEW_TRIAGED,
    'time-sensitive': ItemView.ITEM_VIEW_TIME_SENSITIVE,
    'done': ItemView.ITEM_VIEW_DONE,
  };

  /// Returns the intent that launched the app, if any.
  Future<AppIntent?> getInitialIntent() => _channel.getInitialIntent();

  /// Calls [onIntent] for every intent received while the app is running.
  void listen(void Function(AppIntent) onIntent) => _channel.listen(onIntent);

  void dispose() => _channel.dispose();

  /// Converts the map sent by MainActivity into an [AppIntent].
  static AppIntent? toAppIntent(Map<String, String?> raw) {
    switch (raw['action']) {
      case 'send':
      case 'process_text':
        return parseSharedText(raw['text'] ?? '', subject: raw['subject']);
      case 'view':
        return parseDeepLink(raw['uri'] ?? '');
      default:
        return null;
    }
  }

  /// Parses deep links such as
  /// `todo://item/new?title=Buy%20milk&description=2L&labels=home,errand`,
  /// `todo://items?view=untriaged` and `todo://item/42`.
  static AppIntent? parseDeepLink(String link) {
    final uri = Uri.tryParse(link);
    if (uri == null || uri.scheme != deepLinkScheme) return null;

    final segments = uri.pathSegments.where((s) => s.isNotEmpty).toList();
    switch (uri.host) {
      case 'item':
        if (segments.length != 1) return null;
        if (segments.first == 'new') {
          final params = uri.queryParameters;
          return CreateItemIntent(
            title: nonEmpty(params['title']),
            description: nonEmpty(params['description']),
            labels: splitList(params['labels']),
          );
        }
        final id = int.tryParse(segments.first);
        return id == null || id <= 0 ? null : OpenItemIntent(id);
      case 'items':
        if (segments.isNotEmpty) return null;
        final view = uri.queryParameters['view']?.trim().toLowerCase();
        return OpenViewIntent(_views[view] ?? ItemView.ITEM_VIEW_TRIAGED);
      default:
        return null;
    }
  }

  /// Turns shared or selected text into a new item.
  ///
  /// When [subject] is given (browsers share a page as subject = page title,
  /// text = URL) it becomes the title and [text] the description. Otherwise
  /// the first non-blank line is the title and the remaining lines the
  /// description. A lone URL is kept as the description with its host as the
  /// title.
  static CreateItemIntent parseSharedText(String text, {String? subject}) {
    final trimmed = text.trim();
    final subjectTitle = nonEmpty(subject);
    if (subjectTitle != null) {
      final (title, overflow) = _splitTitle(squashWhitespace(subjectTitle));
      return CreateItemIntent(
        title: title,
        description: nonEmpty(
          [?overflow, if (trimmed.isNotEmpty) trimmed].join('\n\n'),
        ),
      );
    }
    if (trimmed.isEmpty) return const CreateItemIntent();

    final uri = Uri.tryParse(trimmed);
    if (!trimmed.contains(RegExp(r'\s')) &&
        uri != null &&
        uri.hasScheme &&
        uri.host.isNotEmpty) {
      return CreateItemIntent(title: uri.host, description: trimmed);
    }

    final newline = trimmed.indexOf('\n');
    final firstLine = newline < 0 ? trimmed : trimmed.substring(0, newline);
    final rest = newline < 0 ? '' : trimmed.substring(newline + 1).trim();
    final (title, overflow) = _splitTitle(squashWhitespace(firstLine));
    return CreateItemIntent(
      title: title,
      description: nonEmpty(
        [?overflow, if (rest.isNotEmpty) rest].join('\n\n'),
      ),
    );
  }

  /// Splits [line] into a title of at most [maxTitleLength] characters, cut at
  /// the last space when possible, and the overflow (null if none).
  static (String, String?) _splitTitle(String line) {
    if (line.length <= maxTitleLength) return (line, null);
    var cut = line.lastIndexOf(' ', maxTitleLength);
    if (cut <= 0) cut = maxTitleLength;
    return (line.substring(0, cut).trim(), line.substring(cut).trim());
  }
}
