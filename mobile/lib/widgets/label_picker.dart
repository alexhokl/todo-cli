import 'package:flutter/material.dart';

import 'package:todo/l10n/app_localizations.dart';
import 'package:todo/proto/item.pb.dart';
import 'package:todo/widgets/colour.dart';

/// Shared helpers for rendering label chips and the add-label picker dialog.
///
/// Both [ItemDetailPage] (attaching labels to an existing item via
/// `ItemService.updateItemLabels`) and [EditItemPage] (accumulating labels
/// locally and passing them to `ItemService.createItem`) use these helpers so
/// the chip styling, colour swatch, and picker dialog stay consistent.

/// Renders a small colour swatch for [label], falling back to a neutral
/// outline colour when [label.colour] is empty or malformed.
Widget labelColourSwatch(BuildContext context, Label label) {
  if (label.colour.isNotEmpty) {
    final color = parseLabelColour(label.colour);
    if (color != null) {
      return CircleAvatar(backgroundColor: color, maxRadius: 8);
    }
  }
  return CircleAvatar(
    backgroundColor: Theme.of(context).colorScheme.outlineVariant,
    maxRadius: 8,
  );
}

/// Renders [label] as an [InputChip] with a delete affordance wired to
/// [onDeleted]. The chip's avatar is a colour swatch derived from
/// [label.colour] (see [labelColourSwatch]).
Widget labelChip(
  BuildContext context,
  Label label, {
  required VoidCallback onDeleted,
}) {
  final l10n = AppLocalizations.of(context)!;
  Widget? avatar;
  if (label.colour.isNotEmpty) {
    final color = parseLabelColour(label.colour);
    if (color != null) {
      avatar = CircleAvatar(backgroundColor: color, maxRadius: 6);
    }
  }
  return InputChip(
    label: Text(label.name),
    avatar: avatar,
    onDeleted: onDeleted,
    deleteIcon: Icon(Icons.close, semanticLabel: l10n.removeLabel),
  );
}

/// Snapshot of the label catalogue returned by [ensureCatalogue].
class LabelCatalogue {
  const LabelCatalogue({this.labels, this.error});

  /// All known labels, or `null` when no successful load has happened yet.
  final List<Label>? labels;

  /// The most recent catalogue load error, or `null` when the catalogue is
  /// available.
  final String? error;
}

/// Outcome of [showLabelPickerDialog].
class LabelPickerResult {
  LabelPickerResult._({this.label, this.error});

  factory LabelPickerResult.selected(Label? label) =>
      LabelPickerResult._(label: label);

  factory LabelPickerResult.aborted(String error) =>
      LabelPickerResult._(error: error);

  /// The label the user selected, or `null` if the dialog was dismissed
  /// without a selection.
  final Label? label;

  /// Non-null when the catalogue could not be loaded and the dialog was
  /// aborted before showing any options. Callers should surface this as a
  /// SnackBar error.
  final String? error;

  /// True when the catalogue load failed and the dialog was aborted.
  bool get aborted => error != null;

  /// True when the user selected a label (i.e. [label] is non-null).
  bool get hasSelection => label != null;
}

/// Opens a dialog listing known labels not present in [excludedNames] and
/// returns the user's selection, or `null` if dismissed.
///
/// [ensureCatalogue] reloads the label catalogue lazily (when it is missing
/// or a previous load failed) and returns the current snapshot. The caller
/// owns the catalogue state -- typically stored as `List<Label>? _allLabels`
/// and `String? _labelsError` -- so the shared dialog stays stateless.
Future<LabelPickerResult> showLabelPickerDialog(
  BuildContext context, {
  required Future<LabelCatalogue> Function() ensureCatalogue,
  required Set<String> excludedNames,
}) async {
  final l10n = AppLocalizations.of(context)!;

  final catalogue = await ensureCatalogue();
  if (!context.mounted) return LabelPickerResult.aborted('unmounted');

  if (catalogue.error != null || catalogue.labels == null) {
    return LabelPickerResult.aborted(catalogue.error ?? 'labels unavailable');
  }

  final candidates = catalogue.labels!
      .where((l) => !excludedNames.contains(l.name))
      .toList();

  final selected = await showDialog<Label>(
    context: context,
    builder: (context) {
      if (candidates.isEmpty) {
        return SimpleDialog(
          title: Text(l10n.addLabel),
          children: [
            Padding(
              padding:
                  const EdgeInsets.symmetric(horizontal: 24, vertical: 8),
              child: Text(
                l10n.noMoreLabels,
                style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                      color: Theme.of(context)
                          .colorScheme
                          .onSurfaceVariant
                          .withValues(alpha: 0.6),
                      fontStyle: FontStyle.italic,
                    ),
              ),
            ),
          ],
        );
      }
      return SimpleDialog(
        title: Text(l10n.addLabel),
        children: [
          for (final label in candidates)
            SimpleDialogOption(
              onPressed: () => Navigator.of(context).pop(label),
              child: Row(
                children: [
                  labelColourSwatch(context, label),
                  const SizedBox(width: 12),
                  Expanded(child: Text(label.name)),
                ],
              ),
            ),
        ],
      );
    },
  );

  return LabelPickerResult.selected(selected);
}