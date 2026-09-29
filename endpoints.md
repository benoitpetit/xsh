# Endpoints et indicateurs de fonctionnalités X

## Portée du relevé

- Cache local issu de la découverte authentifiée de X : `~/.config/xsh/graphql_ops.json`, horodaté **2026-09-29T11:47:21.106189933+02:00**. Version de cache `1.0`, empreinte de la page `670e31fdb42db122`.
- **104 opérations GraphQL dynamiques actives**, **31 opérations actives disponibles uniquement par repli statique**, soit **135 opérations proposées** par `xsh endpoints list`. Le code contient 63 replis statiques, dont 31 recouvrent une opération dynamique.
- **2 opérations en quarantaine** : `Followers` et `AudioSpaceSearch`. Leurs requêtes réelles ont reçu HTTP 404 ; leurs identifiants sont exclus de la résolution.
- **1103 indicateurs** extraits de la page (594 à `true`, 509 à `false`). **105 entrées** de correspondance opération → indicateurs, dont 62 listes renseignées.

Ce document est un **instantané**, pas une liste codée en dur pour les requêtes. Les identifiants et les valeurs des indicateurs changent côté X. La découverte d’un identifiant dans JavaScript ne prouve pas que l’opération fonctionne. Une opération est dite vérifiée uniquement après une réponse réelle contenant `data` ou après une commande CLI réussie. Aucun cookie ni jeton de session ne figure ici.

## Découverte, résolution et vérification

- La découverte charge une page X authentifiée (`https://x.com/home`, avec repli sur `https://x.com/`), suit les bundles JavaScript référencés, extrait les couples identifiant/nom d’opération et les indicateurs du HTML. La page doit être réellement authentifiée : un shell déconnecté est rejeté.
- Les requêtes GraphQL utilisent `https://x.com/i/api/graphql/{identifiant}/{opération}`. La résolution privilégie le cache dynamique utilisable, puis les replis dans `core/constants.go`. Une opération en quarantaine est retirée de la liste active et son identifiant rejeté ne doit pas être réintroduit par une nouvelle extraction identique.
- Le cache devient ancien après 24 heures et reste utilisable au plus 7 jours ; ces seuils ne prouvent pas la validité des identifiants.
- Une sonde de lecture sûre existe pour `HomeTimeline` et `HomeLatestTimeline` (GET), `UserByScreenName` (GET), `Viewer` (GET), `ExplorePage` (GET), `AudioSpaceSearch` (GET, requête limitée à un résultat), et `SearchTimeline` (POST). Elle exige une réponse GraphQL avec `data`. Pour les autres opérations, `xsh endpoints check <opération>` renvoie `unsupported` si aucun jeu de variables sûr n’est défini. Les mutations ne sont pas exécutées comme sondes.
- `xsh endpoints check <opération>` sépare l’état de découverte (`discovered`, `static_fallback`, `quarantined`, `unavailable`) de la vérification en direct (`healthy`, `obsolete`, `auth_error`, `inconclusive`, `unsupported`, `unavailable`). `verified` n’est vrai qu’après une réponse contenant `data`. La sortie indique aussi si les indicateurs viennent de la correspondance propre à l’opération ou des valeurs intégrées par défaut.
- Une réponse `obsolete` confirmée par `xsh endpoints check` est enregistrée en quarantaine avec l’identifiant rejeté. La commande retire cet endpoint de la sélection ; une nouvelle découverte peut le réactiver si X publie un identifiant différent.
- `xsh endpoints status` et `xsh status --check` signalent aussi les opérations en quarantaine ; le statut global peut donc être dégradé alors que les sondes critiques réussissent.

### Résultats réels observés

| Opération | Résultat observé | Portée |
| --- | --- | --- |
| `UserByScreenName` | Réponse GraphQL contenant `data` | Identifiant dynamique |
| `SearchTimeline` | Réponse GraphQL contenant `data` | Identifiant dynamique |
| `Viewer` | Réponse GraphQL contenant `data` | Identifiant dynamique |
| `ExplorePage` | Réponse GraphQL contenant `data` | Identifiant dynamique |
| `HomeTimeline` | Réponse GraphQL contenant `data` | Repli statique |
| `HomeLatestTimeline` | Réponse GraphQL contenant `data` | Repli statique |
| `Following` | Commande de lecture CLI réussie | Parcours CLI, pas une sonde générique |
| `Followers` | HTTP 404 lors de `xsh user followers devbyben --count 1 --account default --json` | Identifiant découvert puis mis en quarantaine |
| `AudioSpaceSearch` | HTTP 404 via `xsh endpoints check`; l’identifiant de repli est maintenant en quarantaine | Repli statique retiré de la sélection |

Les états ci-dessus viennent des sondes exécutées le 29 septembre 2026 ; ils ne sont pas revérifiés en continu. `xsh auth whoami --account default --json` a également résolu le profil et renvoyé une image de profil non vide.

## Inventaire GraphQL

`Dynamique` signifie extrait des bundles et actif dans le cache ; `Repli` signifie défini dans `core/constants.go`. La colonne « Vérification » reste `non sondé` pour les opérations seulement extraites. Les chemins ci-dessous sont relatifs à `https://x.com/i/api/graphql/`.

| Opération | Chemin | Origine | Vérification |
| --- | --- | --- | --- |
| `AddContentDisclosure` | `D1nwFlsu_qHsX92YzoRaaA/AddContentDisclosure` | Dynamique | non sondé |
| `AudioSpaceById` | `Uv5R_x2bBDPMVxfGjkzxtA/AudioSpaceById` | Repli | non sondé |
| `AudioSpaceSearch` | `NTq79TuSz6fHj8lQaferJw/AudioSpaceSearch` | Repli (quarantaine) | HTTP 404 confirmé |
| `AuthenticatePeriscope` | `r7VUmxbfqNkx7uwjgONSNw/AuthenticatePeriscope` | Dynamique | non sondé |
| `BakeryQuery` | `pROR-yRiBVsEjJyHt3fvhg/BakeryQuery` | Dynamique | non sondé |
| `BlockedAccountsAll` | `17j54yk-VPnOed6s1EdosQ/BlockedAccountsAll` | Dynamique + repli | non sondé |
| `BlockedAccountsImported` | `NOopxUl-dcnnOI1_zXQyLg/BlockedAccountsImported` | Dynamique | non sondé |
| `BlueVerifiedFollowers` | `ck_SV_kTAlbD2WZiOFNbzw/BlueVerifiedFollowers` | Dynamique + repli | non sondé |
| `BookmarkFoldersSlice` | `i78YDd0Tza-dV4SYs58kRg/BookmarkFoldersSlice` | Repli | non sondé |
| `BookmarkFolderTimeline` | `hNY7X2xE2N7HVF6Qb_mu6w/BookmarkFolderTimeline` | Repli | non sondé |
| `Bookmarks` | `uKP9v_I31k0_VSBmlpq2Xg/Bookmarks` | Repli | non sondé |
| `BookmarkSearchTimeline` | `vqbH512y_Emfr0P4rXJ0Lw/BookmarkSearchTimeline` | Dynamique + repli | non sondé |
| `CommunitiesMainPageTimeline` | `ZYMHvNRSCYjfaxcK_XBIOQ/CommunitiesMainPageTimeline` | Repli | non sondé |
| `CommunityByRestId` | `nNjM4SaF_W1Hk0JQ8pOPQg/CommunityByRestId` | Repli | non sondé |
| `CommunityMemberRelationshipTypeahead` | `6YgvBKI7c3YZ9d7zKKojng/CommunityMemberRelationshipTypeahead` | Dynamique | non sondé |
| `CommunityTweetsTimeline` | `aeJsV55JMPw15_-N4VCnPA/CommunityTweetsTimeline` | Repli | non sondé |
| `CommunityUserRelationshipTypeahead` | `MOKE3VjNMJ7BVsWdNkv32g/CommunityUserRelationshipTypeahead` | Dynamique | non sondé |
| `ConnectTabTimeline` | `Chu4A5yjQn8BVizIpzQaQA/ConnectTabTimeline` | Dynamique | non sondé |
| `ConversationControlChange` | `57WYJNnWH0vM3Ip_gm8B2g/ConversationControlChange` | Dynamique | non sondé |
| `ConversationControlDelete` | `OoMO_aSZ1ZXjegeamF9QmA/ConversationControlDelete` | Dynamique | non sondé |
| `CreateBookmark` | `aoDbu3RHznuiSkQ9aNM67Q/CreateBookmark` | Dynamique + repli | non sondé |
| `CreateHighlight` | `7jEc7ECTTDcNaqsMhjTxXg/CreateHighlight` | Dynamique | non sondé |
| `CreateList` | `QXil-VE8uEJPfUKFiO36Bg/CreateList` | Repli | non sondé |
| `CreateNoteTweet` | `Q-sJyowqllPknxuKDt_pKQ/CreateNoteTweet` | Dynamique + repli | non sondé |
| `CreateRetweet` | `mbRO74GrOvSfRcJnlMapnQ/CreateRetweet` | Dynamique + repli | non sondé |
| `CreateScheduledTweet` | `LCVzRQGxOaGnOnYH01NQXg/CreateScheduledTweet` | Repli | non sondé |
| `CreateTweet` | `WNkbkQ_JLIofjdukTXahVA/CreateTweet` | Dynamique + repli | non sondé |
| `DataSaverMode` | `xF6sXnKJfS2AOylzxRjf6A/DataSaverMode` | Dynamique | non sondé |
| `DeleteBookmark` | `Wlmlj2-xzyS1GN3a6cj-mQ/DeleteBookmark` | Dynamique + repli | non sondé |
| `DeleteContentDisclosure` | `YeIV-eqGwEZXDtYaDsJz2Q/DeleteContentDisclosure` | Dynamique | non sondé |
| `DeleteHighlight` | `ea-VVDSLIEYNY2_2aPg3Uw/DeleteHighlight` | Dynamique | non sondé |
| `DeleteList` | `UnN9Th1BDbeLjpgjGSpL3Q/DeleteList` | Repli | non sondé |
| `DeleteRetweet` | `ZyZigVsNiFO6v1dEks1eWg/DeleteRetweet` | Dynamique + repli | non sondé |
| `DeleteScheduledTweet` | `CTOVqej0JBXAZSwkp1US0g/DeleteScheduledTweet` | Repli | non sondé |
| `DeleteTweet` | `nxpZCY2K-I6QoFHAHeojFQ/DeleteTweet` | Dynamique + repli | non sondé |
| `DisableVerifiedPhoneLabel` | `g2m0pAOamawNtVIfjXNMJg/DisableVerifiedPhoneLabel` | Dynamique | non sondé |
| `dmBlockUser` | `IYw9u1KEhrS-t-BXsau4Uw/dmBlockUser` | Dynamique | non sondé |
| `DMMessageDeleteMutation` | `BJ6DtxA2llfjnRoRjaiIiw/DMMessageDeleteMutation` | Repli | non sondé |
| `DmNsfwMediaFilterUpdate` | `of_N6O33zfyD4qsFJMYFxA/DmNsfwMediaFilterUpdate` | Dynamique | non sondé |
| `dmUnblockUser` | `Krbs6Nak_o7liWQwfV1jOQ/dmUnblockUser` | Dynamique | non sondé |
| `DownvoteTweet` | `Iu4kUV4vd_iHMupiHPPrAQ/DownvoteTweet` | Dynamique | non sondé |
| `EnableVerifiedPhoneLabel` | `C3RJFfMsb_KcEytpKmRRkw/EnableVerifiedPhoneLabel` | Dynamique | non sondé |
| `ExplorePage` | `e9FxSelvGhzaVvCyfmd-HA/ExplorePage` | Dynamique + repli | non sondé |
| `ExploreSidebar` | `a_EaOYRmowjytmooCbPRtA/ExploreSidebar` | Dynamique | non sondé |
| `FavoriteTweet` | `lI07N6Otwv1PhnEgXILM7A/FavoriteTweet` | Dynamique + repli | non sondé |
| `FetchScheduledTweets` | `ITtjAzvlZni2wWXwf295Qg/FetchScheduledTweets` | Repli | non sondé |
| `FinanceSearchTags` | `4VTsYZ3r4xYW9S0Rys5G0A/FinanceSearchTags` | Dynamique | non sondé |
| `FollowersYouKnow` | `kSjQs8VV3c9WKxUoutJcrw/FollowersYouKnow` | Dynamique + repli | non sondé |
| `Following` | `uwmIAx89XrXNuGY-Y7WFLg/Following` | Dynamique + repli | Lecture CLI réussie |
| `FollowUser` | `iNZ4xPly3JveJqEshzJdLA/FollowUser` | Repli | non sondé |
| `GenericTimelineById` | `S_hzVUv1trgZ_5ruDe2IoA/GenericTimelineById` | Dynamique | non sondé |
| `getAltTextPromptPreference` | `PFIxTk8owMoZgiMccP0r4g/getAltTextPromptPreference` | Dynamique | non sondé |
| `GetUserClaims` | `aQ-b88K_Lp7dgHX53MqNQQ/GetUserClaims` | Dynamique | non sondé |
| `GetUsernameAvailabilityAndSuggestions` | `1bMz-9lPrmIXrhFmXntTHw/GetUsernameAvailabilityAndSuggestions` | Dynamique | non sondé |
| `GlobalCommunitiesLatestPostSearchTimeline` | `c2zrxcOVSiQFzV1LvnKYXg/GlobalCommunitiesLatestPostSearchTimeline` | Dynamique | non sondé |
| `GlobalCommunitiesPostSearchTimeline` | `vcCW9agd-RJnwBZrxVkzFg/GlobalCommunitiesPostSearchTimeline` | Dynamique | non sondé |
| `HomeLatestTimeline` | `JVzDMxTXbT9bRXSpUR16CQ/HomeLatestTimeline` | Repli | non sondé |
| `HomeTimeline` | `gXtpuBkna6SRLFFKaT2OTg/HomeTimeline` | Repli | GraphQL `data` |
| `JobScreenQuery` | `8uZH_OBKTFNIMzTJaV5lbQ/JobScreenQuery` | Repli | non sondé |
| `JobSearchQueryScreenJobsQuery` | `jVMK9qcOUB5xQQdSLr5ECg/JobSearchQueryScreenJobsQuery` | Repli | non sondé |
| `JoinCommunity` | `bTM1mhHSML-sKB9EjT0RRg/JoinCommunity` | Repli | non sondé |
| `LeaveCommunity` | `OoS6Nd__MZ0LHAiBbmc3gQ/LeaveCommunity` | Repli | non sondé |
| `Likes` | `PgAssYGsPMMF1vVox5ysPg/Likes` | Dynamique + repli | non sondé |
| `ListAddMember` | `nAi8BAjn1xQOyCH0hWZpPA/ListAddMember` | Repli | non sondé |
| `ListByRestId` | `bSE1lqLBnovM86uu4p4Iqg/ListByRestId` | Repli | non sondé |
| `ListLatestTweetsTimeline` | `gNXkRRRV67cSRJkmpgGPnA/ListLatestTweetsTimeline` | Repli | non sondé |
| `ListMembers` | `fqecRWCF4EcSAOs5yXh7Ig/ListMembers` | Repli | non sondé |
| `ListMemberships` | `cRsrj8HASXYzxaf90wMDPQ/ListMemberships` | Repli | non sondé |
| `ListRemoveMember` | `pGMiwtWRMx08r4XCYxai4Q/ListRemoveMember` | Repli | non sondé |
| `ListSearchTimeline` | `6RIvCXM8i3HnC2_Rshgb_w/ListSearchTimeline` | Dynamique | non sondé |
| `ListsManagementPageTimeline` | `qXWhKTaNeJianB02_JovCg/ListsManagementPageTimeline` | Repli | non sondé |
| `MediaTabVideoMixer` | `SJlH7KNABwCu-8BxJGIqzw/MediaTabVideoMixer` | Dynamique | non sondé |
| `ModeratedTimeline` | `Q51KPb_uAJZDWhsvRcKToQ/ModeratedTimeline` | Dynamique | non sondé |
| `ModerateTweet` | `pjFnHGVqCjTcZol0xcBJjw/ModerateTweet` | Dynamique | non sondé |
| `MutedAccounts` | `SkK3NXvsAGlAlcNhRBDiKw/MutedAccounts` | Dynamique + repli | non sondé |
| `PaymentsUsersTypeahead` | `dlHFEO5q9vveDgITz9zELg/PaymentsUsersTypeahead` | Dynamique | non sondé |
| `PinnableTimelines` | `XUpSmUE_n6ez-U_SdC7G7g/PinnableTimelines` | Dynamique | non sondé |
| `PinnedTimelines` | `1C9qXYjxcujNpyJWE6tAeg/PinnedTimelines` | Dynamique | non sondé |
| `PinTimeline` | `ZMYcJPUQ0QIAyWekMdAuDw/PinTimeline` | Dynamique + repli | non sondé |
| `PinTweet` | `VIHsNu89pK-kW35JpHq7Xw/PinTweet` | Dynamique | non sondé |
| `ProfileFilter` | `bIYQNt7PiYua6tWDAuTY3Q/ProfileFilter` | Dynamique | non sondé |
| `ProfileUserPhoneState` | `5kUWP8C1hcd6omvg6HXXTQ/ProfileUserPhoneState` | Dynamique | non sondé |
| `PutClientEducationFlag` | `IjQ-egg0uPkY11NyPMfRMQ/PutClientEducationFlag` | Dynamique | non sondé |
| `RemoveFollower` | `QpNfg0kpPRfjROQ_9eOLXA/RemoveFollower` | Dynamique | non sondé |
| `SearchTimeline` | `uGB-gNd5HE4TkpO70OcFNw/SearchTimeline` | Dynamique + repli | GraphQL `data` |
| `SharingAudiospacesListeningDataWithFollowersUpdate` | `5h0kNbk3ii97rmfY6CdgAA/SharingAudiospacesListeningDataWithFollowersUpdate` | Dynamique | non sondé |
| `SimilarPosts` | `7DS57_pJ2xS39b3yBA6xew/SimilarPosts` | Dynamique | non sondé |
| `SuperFollowers` | `Gc9oHdW7jb_9tbxHCR5vGA/SuperFollowers` | Dynamique | non sondé |
| `SupportedLanguages` | `fZ5uZVeledO5SAseKnmTUg/SupportedLanguages` | Dynamique | non sondé |
| `timelinesFeedback` | `vfVbgvTPTQ-dF_PQ5lD1WQ/timelinesFeedback` | Dynamique | non sondé |
| `TrendHistory` | `Livn0q_jXpeC5EVqxgu4Sg/TrendHistory` | Dynamique | non sondé |
| `TrendRelevantUsers` | `dY98yXH0SaJxntCw4Bqrfg/TrendRelevantUsers` | Dynamique | non sondé |
| `Trends` | `Z6s1tFEq4BveGOj0N80z8g/ExplorePage` | Repli | non sondé |
| `TVHomeMixer` | `pj3yJUnDE8vmJu5k-W9OSA/TVHomeMixer` | Dynamique | non sondé |
| `TweetDetail` | `blErEeZkos5TDrWmrCp7cw/TweetDetail` | Dynamique + repli | non sondé |
| `TweetQuotes` | `kv6oBLVLyPi0JWVqrY4kJQ/TweetQuotes` | Repli | non sondé |
| `TweetResultByRestId` | `LbQZrAWyKPvExi8di3-EoA/TweetResultByRestId` | Dynamique + repli | non sondé |
| `TweetResultsByRestIds` | `RRYnxFsEuhlm9c0lWHDgaQ/TweetResultsByRestIds` | Dynamique + repli | non sondé |
| `UndoDownvoteTweet` | `yqhcbdyy59k-FCwOysvvGQ/UndoDownvoteTweet` | Dynamique | non sondé |
| `UnfavoriteTweet` | `ZYKSe-w7KEslx3JhSIk5LA/UnfavoriteTweet` | Dynamique + repli | non sondé |
| `UnfollowUser` | `tTT1x7v4h9zLVzT1BA/UnfollowUser` | Repli | non sondé |
| `UnmentionUserFromConversation` | `xVW9j3OqoBRY9d6_2OONEg/UnmentionUserFromConversation` | Dynamique | non sondé |
| `UnmoderateTweet` | `pVSyu6PA57TLvIE4nN2tsA/UnmoderateTweet` | Dynamique | non sondé |
| `UnpinTimeline` | `_uu6eZuimGwywEa4qtZxZA/UnpinTimeline` | Dynamique + repli | non sondé |
| `UnpinTweet` | `BhKei844ypCyLYCg0nwigw/UnpinTweet` | Dynamique | non sondé |
| `updateAltTextPromptPreference` | `aQKrduk_DA46XfOQDkcEng/updateAltTextPromptPreference` | Dynamique | non sondé |
| `UpdateList` | `qE2QVWL84jqa6CmH-m-D3w/UpdateList` | Repli | non sondé |
| `UpdatePinnedTimelines` | `AtN-0mKI3fXXmxzYYk1Wqw/UpdatePinnedTimelines` | Dynamique | non sondé |
| `Upsells` | `Sg3BvwapuCMIjLJ7LGPhMA/Upsells` | Dynamique | non sondé |
| `UrtFixtures` | `S1l5Sr3QVIsc2HGWhowrmA/UrtFixtures` | Dynamique | non sondé |
| `UserArticlesTweets` | `H7U-Vk99Mgbgl1fqS_kHmA/UserArticlesTweets` | Dynamique | non sondé |
| `UserBusinessProfileTeamTimeline` | `KmCGUSqQhyectU4dXrKABQ/UserBusinessProfileTeamTimeline` | Dynamique | non sondé |
| `UserByRestId` | `IdmRdjYxIGI39Hdwkwo5cQ/UserByRestId` | Dynamique + repli | non sondé |
| `UserByScreenName` | `KybxDj9RrADIITXlGG8kpw/UserByScreenName` | Dynamique + repli | GraphQL `data` |
| `UserCreatorSubscribers` | `Rq1DG2wcHXZqszcUuI1_UQ/UserCreatorSubscribers` | Dynamique | non sondé |
| `UserCreatorSubscriptions` | `TIP2yqAkV5plyDLkVKYDFw/UserCreatorSubscriptions` | Dynamique | non sondé |
| `UserHighlightsTweets` | `kzlhnAZg6za8KVcUlitqBQ/UserHighlightsTweets` | Dynamique | non sondé |
| `UserMedia` | `GEs4r5bWKm0P0EIRfo2DGw/UserMedia` | Dynamique + repli | non sondé |
| `UserOriginalsTimeline` | `qtvmQffnepvr0oPe4A8MqQ/UserOriginalsTimeline` | Dynamique | non sondé |
| `UserPhotoTimeline` | `YqEBDpaXbWuRPks59hau0g/UserPhotoTimeline` | Dynamique | non sondé |
| `UserPreferences` | `xFxU-O8hEYe74ovNVU74jA/UserPreferences` | Dynamique | non sondé |
| `UserPromotableTweets` | `90MFDSfkTYifH-Y9uK0RRA/UserPromotableTweets` | Dynamique | non sondé |
| `UserPromotedTweets` | `LfWVX0LfihfWaO89eTkmmA/UserPromotedTweets` | Dynamique | non sondé |
| `UserRepliesTimeline` | `9FLI4sKKO6rEojPOEHT7BA/UserRepliesTimeline` | Dynamique | non sondé |
| `UserRepostsTimeline` | `hkQQA_PMJfzHlRtnUYYXMg/UserRepostsTimeline` | Dynamique | non sondé |
| `UsersByRestIds` | `BuQFwM7wpHl00cfHL-r0rA/UsersByRestIds` | Dynamique + repli | non sondé |
| `UsersByScreenNames` | `8G9O4pAkWTXNv3XQuOGVrw/UsersByScreenNames` | Dynamique + repli | non sondé |
| `UserSessionsList` | `vJ-XatpmQSG8bDch8-t9Jw/UserSessionsList` | Dynamique | non sondé |
| `UserSuperFollowTweets` | `UAtqQcjjYTU-SOp1ZIoWPw/UserSuperFollowTweets` | Dynamique | non sondé |
| `UsersVerifiedAvatars` | `O9O_O1iF0QffoOBjB2OGIQ/UsersVerifiedAvatars` | Dynamique | non sondé |
| `UserTweets` | `qJy3MbaNndtzxf9IqUzxMg/UserTweets` | Dynamique + repli | non sondé |
| `UserTweetsAndReplies` | `Z1m9j8S1leAzQp6yZXuaSg/UserTweetsAndReplies` | Dynamique + repli | non sondé |
| `UserVideoTimeline` | `5A9PzD08T6PbvC2QlEYxMg/UserVideoTimeline` | Dynamique | non sondé |
| `Viewer` | `9t128XgFic52jPUEkJMf6w/Viewer` | Dynamique + repli | non sondé |
| `ViewerBadgeCounts` | `q4Npr1-FYRWyXsRzPckwEA/ViewerBadgeCounts` | Dynamique | non sondé |
| `WriteDataSaverPreferences` | `H03etWvZGz41YASxAU2YPg/WriteDataSaverPreferences` | Dynamique | non sondé |

### Opérations exclues

| Opération | Identifiant rejeté | Motif |
| --- | --- | --- |
| `AudioSpaceSearch` | `NTq79TuSz6fHj8lQaferJw/AudioSpaceSearch` | GraphQL operation returned HTTP 404 |
| `Followers` | `mrqxgX8JzwlL6pvYiC5CPA/Followers` | HTTP 404 or query not found |

`Trends` est un alias statique vers le chemin `ExplorePage`. Il ne correspond pas à une opération GraphQL distincte dans ce repli.

## Routes REST connues du code

Ces routes sont codées dans la CLI ; elles ne proviennent pas du relevé des bundles GraphQL. Leur présence dans le code ne constitue pas une vérification réseau actuelle.

| Méthode | Route | Usage dans la CLI |
| --- | --- | --- |
| GET | `https://x.com/i/api/2/notifications/all.json` | Notifications |
| GET | `https://x.com/i/api/1.1/dm/inbox_initial_state.json` | Boîte de messages privés |
| POST | `https://x.com/i/api/1.1/dm/new2.json` | Envoyer un message privé |
| POST | `https://x.com/i/api/1.1/friendships/create.json` | Suivre |
| POST | `https://x.com/i/api/1.1/friendships/destroy.json` | Ne plus suivre |
| POST | `https://x.com/i/api/1.1/blocks/create.json` | Bloquer |
| POST | `https://x.com/i/api/1.1/blocks/destroy.json` | Débloquer |
| POST | `https://x.com/i/api/1.1/mutes/users/create.json` | Masquer |
| POST | `https://x.com/i/api/1.1/mutes/users/destroy.json` | Annuler le masquage |
| POST | `https://upload.twitter.com/i/media/upload.json` | Téléverser un média |

## Correspondances opération → indicateurs

La correspondance est extraite des bundles. `—` indique que le cache n’a pas fourni de liste pour cette opération ; cela ne signifie pas qu’elle n’utilise aucun indicateur. En l’absence de liste, la CLI utilise ses valeurs par défaut. Pour une liste renseignée, elle prend la valeur globale du cache et utilise `true` si la clé manque.

| Opération | Nombre | Indicateurs référencés |
| --- | ---: | --- |
| `AddContentDisclosure` | 0 | — |
| `AuthenticatePeriscope` | 0 | — |
| `BakeryQuery` | 0 | — |
| `BlockedAccountsAll` | 40 | `rweb_video_screen_enabled`, `rweb_cashtags_enabled`, `profile_label_improvements_pcf_label_in_post_enabled`, `responsive_web_profile_redirect_enabled`, `rweb_tipjar_consumption_enabled`, `verified_phone_label_enabled`, `creator_subscriptions_tweet_preview_api_enabled`, `responsive_web_graphql_timeline_navigation_enabled`, `premium_content_api_read_enabled`, `communities_web_enable_tweet_community_results_fetch`, `c9s_tweet_anatomy_moderator_badge_enabled`, `responsive_web_grok_analyze_button_fetch_trends_enabled`, `responsive_web_grok_analyze_post_followups_enabled`, `rweb_cashtags_composer_attachment_enabled`, `responsive_web_jetfuel_frame`, `rweb_sports_post_context_enabled`, `responsive_web_grok_share_attachment_enabled`, `responsive_web_grok_annotations_enabled`, `articles_preview_enabled`, `responsive_web_edit_tweet_api_enabled`, `rweb_conversational_replies_downvote_enabled`, `graphql_is_translatable_rweb_tweet_is_translatable_enabled`, `view_counts_everywhere_api_enabled`, `longform_notetweets_consumption_enabled`, `responsive_web_twitter_article_tweet_consumption_enabled`, `content_disclosure_indicator_enabled`, `content_disclosure_ai_generated_indicator_enabled`, `responsive_web_grok_show_grok_translated_post`, `responsive_web_grok_analysis_button_from_backend`, `post_ctas_fetch_enabled`, `freedom_of_speech_not_reach_fetch_enabled`, `standardized_nudges_misinfo`, `tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled`, `longform_notetweets_rich_text_read_enabled`, `longform_notetweets_inline_media_enabled`, `responsive_web_nested_quote_preview_enabled`, `responsive_web_grok_image_annotation_enabled`, `responsive_web_grok_imagine_annotation_enabled`, `responsive_web_grok_community_note_auto_translation_is_enabled`, `responsive_web_enhance_cards_enabled` |
| `BlockedAccountsImported` | 40 | `rweb_video_screen_enabled`, `rweb_cashtags_enabled`, `profile_label_improvements_pcf_label_in_post_enabled`, `responsive_web_profile_redirect_enabled`, `rweb_tipjar_consumption_enabled`, `verified_phone_label_enabled`, `creator_subscriptions_tweet_preview_api_enabled`, `responsive_web_graphql_timeline_navigation_enabled`, `premium_content_api_read_enabled`, `communities_web_enable_tweet_community_results_fetch`, `c9s_tweet_anatomy_moderator_badge_enabled`, `responsive_web_grok_analyze_button_fetch_trends_enabled`, `responsive_web_grok_analyze_post_followups_enabled`, `rweb_cashtags_composer_attachment_enabled`, `responsive_web_jetfuel_frame`, `rweb_sports_post_context_enabled`, `responsive_web_grok_share_attachment_enabled`, `responsive_web_grok_annotations_enabled`, `articles_preview_enabled`, `responsive_web_edit_tweet_api_enabled`, `rweb_conversational_replies_downvote_enabled`, `graphql_is_translatable_rweb_tweet_is_translatable_enabled`, `view_counts_everywhere_api_enabled`, `longform_notetweets_consumption_enabled`, `responsive_web_twitter_article_tweet_consumption_enabled`, `content_disclosure_indicator_enabled`, `content_disclosure_ai_generated_indicator_enabled`, `responsive_web_grok_show_grok_translated_post`, `responsive_web_grok_analysis_button_from_backend`, `post_ctas_fetch_enabled`, `freedom_of_speech_not_reach_fetch_enabled`, `standardized_nudges_misinfo`, `tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled`, `longform_notetweets_rich_text_read_enabled`, `longform_notetweets_inline_media_enabled`, `responsive_web_nested_quote_preview_enabled`, `responsive_web_grok_image_annotation_enabled`, `responsive_web_grok_imagine_annotation_enabled`, `responsive_web_grok_community_note_auto_translation_is_enabled`, `responsive_web_enhance_cards_enabled` |
| `BlueVerifiedFollowers` | 40 | `rweb_video_screen_enabled`, `rweb_cashtags_enabled`, `profile_label_improvements_pcf_label_in_post_enabled`, `responsive_web_profile_redirect_enabled`, `rweb_tipjar_consumption_enabled`, `verified_phone_label_enabled`, `creator_subscriptions_tweet_preview_api_enabled`, `responsive_web_graphql_timeline_navigation_enabled`, `premium_content_api_read_enabled`, `communities_web_enable_tweet_community_results_fetch`, `c9s_tweet_anatomy_moderator_badge_enabled`, `responsive_web_grok_analyze_button_fetch_trends_enabled`, `responsive_web_grok_analyze_post_followups_enabled`, `rweb_cashtags_composer_attachment_enabled`, `responsive_web_jetfuel_frame`, `rweb_sports_post_context_enabled`, `responsive_web_grok_share_attachment_enabled`, `responsive_web_grok_annotations_enabled`, `articles_preview_enabled`, `responsive_web_edit_tweet_api_enabled`, `rweb_conversational_replies_downvote_enabled`, `graphql_is_translatable_rweb_tweet_is_translatable_enabled`, `view_counts_everywhere_api_enabled`, `longform_notetweets_consumption_enabled`, `responsive_web_twitter_article_tweet_consumption_enabled`, `content_disclosure_indicator_enabled`, `content_disclosure_ai_generated_indicator_enabled`, `responsive_web_grok_show_grok_translated_post`, `responsive_web_grok_analysis_button_from_backend`, `post_ctas_fetch_enabled`, `freedom_of_speech_not_reach_fetch_enabled`, `standardized_nudges_misinfo`, `tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled`, `longform_notetweets_rich_text_read_enabled`, `longform_notetweets_inline_media_enabled`, `responsive_web_nested_quote_preview_enabled`, `responsive_web_grok_image_annotation_enabled`, `responsive_web_grok_imagine_annotation_enabled`, `responsive_web_grok_community_note_auto_translation_is_enabled`, `responsive_web_enhance_cards_enabled` |
| `BookmarkSearchTimeline` | 40 | `rweb_video_screen_enabled`, `rweb_cashtags_enabled`, `profile_label_improvements_pcf_label_in_post_enabled`, `responsive_web_profile_redirect_enabled`, `rweb_tipjar_consumption_enabled`, `verified_phone_label_enabled`, `creator_subscriptions_tweet_preview_api_enabled`, `responsive_web_graphql_timeline_navigation_enabled`, `premium_content_api_read_enabled`, `communities_web_enable_tweet_community_results_fetch`, `c9s_tweet_anatomy_moderator_badge_enabled`, `responsive_web_grok_analyze_button_fetch_trends_enabled`, `responsive_web_grok_analyze_post_followups_enabled`, `rweb_cashtags_composer_attachment_enabled`, `responsive_web_jetfuel_frame`, `rweb_sports_post_context_enabled`, `responsive_web_grok_share_attachment_enabled`, `responsive_web_grok_annotations_enabled`, `articles_preview_enabled`, `responsive_web_edit_tweet_api_enabled`, `rweb_conversational_replies_downvote_enabled`, `graphql_is_translatable_rweb_tweet_is_translatable_enabled`, `view_counts_everywhere_api_enabled`, `longform_notetweets_consumption_enabled`, `responsive_web_twitter_article_tweet_consumption_enabled`, `content_disclosure_indicator_enabled`, `content_disclosure_ai_generated_indicator_enabled`, `responsive_web_grok_show_grok_translated_post`, `responsive_web_grok_analysis_button_from_backend`, `post_ctas_fetch_enabled`, `freedom_of_speech_not_reach_fetch_enabled`, `standardized_nudges_misinfo`, `tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled`, `longform_notetweets_rich_text_read_enabled`, `longform_notetweets_inline_media_enabled`, `responsive_web_nested_quote_preview_enabled`, `responsive_web_grok_image_annotation_enabled`, `responsive_web_grok_imagine_annotation_enabled`, `responsive_web_grok_community_note_auto_translation_is_enabled`, `responsive_web_enhance_cards_enabled` |
| `CommunityMemberRelationshipTypeahead` | 1 | `responsive_web_graphql_timeline_navigation_enabled` |
| `CommunityUserRelationshipTypeahead` | 1 | `responsive_web_graphql_timeline_navigation_enabled` |
| `ConnectTabTimeline` | 40 | `rweb_video_screen_enabled`, `rweb_cashtags_enabled`, `profile_label_improvements_pcf_label_in_post_enabled`, `responsive_web_profile_redirect_enabled`, `rweb_tipjar_consumption_enabled`, `verified_phone_label_enabled`, `creator_subscriptions_tweet_preview_api_enabled`, `responsive_web_graphql_timeline_navigation_enabled`, `premium_content_api_read_enabled`, `communities_web_enable_tweet_community_results_fetch`, `c9s_tweet_anatomy_moderator_badge_enabled`, `responsive_web_grok_analyze_button_fetch_trends_enabled`, `responsive_web_grok_analyze_post_followups_enabled`, `rweb_cashtags_composer_attachment_enabled`, `responsive_web_jetfuel_frame`, `rweb_sports_post_context_enabled`, `responsive_web_grok_share_attachment_enabled`, `responsive_web_grok_annotations_enabled`, `articles_preview_enabled`, `responsive_web_edit_tweet_api_enabled`, `rweb_conversational_replies_downvote_enabled`, `graphql_is_translatable_rweb_tweet_is_translatable_enabled`, `view_counts_everywhere_api_enabled`, `longform_notetweets_consumption_enabled`, `responsive_web_twitter_article_tweet_consumption_enabled`, `content_disclosure_indicator_enabled`, `content_disclosure_ai_generated_indicator_enabled`, `responsive_web_grok_show_grok_translated_post`, `responsive_web_grok_analysis_button_from_backend`, `post_ctas_fetch_enabled`, `freedom_of_speech_not_reach_fetch_enabled`, `standardized_nudges_misinfo`, `tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled`, `longform_notetweets_rich_text_read_enabled`, `longform_notetweets_inline_media_enabled`, `responsive_web_nested_quote_preview_enabled`, `responsive_web_grok_image_annotation_enabled`, `responsive_web_grok_imagine_annotation_enabled`, `responsive_web_grok_community_note_auto_translation_is_enabled`, `responsive_web_enhance_cards_enabled` |
| `ConversationControlChange` | 0 | — |
| `ConversationControlDelete` | 0 | — |
| `CreateBookmark` | 0 | — |
| `CreateHighlight` | 0 | — |
| `CreateNoteTweet` | 37 | `premium_content_api_read_enabled`, `communities_web_enable_tweet_community_results_fetch`, `c9s_tweet_anatomy_moderator_badge_enabled`, `responsive_web_grok_analyze_button_fetch_trends_enabled`, `responsive_web_grok_analyze_post_followups_enabled`, `rweb_cashtags_composer_attachment_enabled`, `responsive_web_jetfuel_frame`, `rweb_sports_post_context_enabled`, `responsive_web_grok_share_attachment_enabled`, `responsive_web_grok_annotations_enabled`, `responsive_web_edit_tweet_api_enabled`, `rweb_conversational_replies_downvote_enabled`, `graphql_is_translatable_rweb_tweet_is_translatable_enabled`, `view_counts_everywhere_api_enabled`, `longform_notetweets_consumption_enabled`, `responsive_web_twitter_article_tweet_consumption_enabled`, `content_disclosure_indicator_enabled`, `content_disclosure_ai_generated_indicator_enabled`, `responsive_web_grok_show_grok_translated_post`, `responsive_web_grok_analysis_button_from_backend`, `post_ctas_fetch_enabled`, `longform_notetweets_rich_text_read_enabled`, `longform_notetweets_inline_media_enabled`, `profile_label_improvements_pcf_label_in_post_enabled`, `responsive_web_profile_redirect_enabled`, `rweb_tipjar_consumption_enabled`, `verified_phone_label_enabled`, `responsive_web_nested_quote_preview_enabled`, `articles_preview_enabled`, `rweb_cashtags_enabled`, `responsive_web_grok_community_note_auto_translation_is_enabled`, `freedom_of_speech_not_reach_fetch_enabled`, `standardized_nudges_misinfo`, `tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled`, `responsive_web_grok_image_annotation_enabled`, `responsive_web_grok_imagine_annotation_enabled`, `responsive_web_graphql_timeline_navigation_enabled` |
| `CreateRetweet` | 0 | — |
| `CreateTweet` | 37 | `premium_content_api_read_enabled`, `communities_web_enable_tweet_community_results_fetch`, `c9s_tweet_anatomy_moderator_badge_enabled`, `responsive_web_grok_analyze_button_fetch_trends_enabled`, `responsive_web_grok_analyze_post_followups_enabled`, `rweb_cashtags_composer_attachment_enabled`, `responsive_web_jetfuel_frame`, `rweb_sports_post_context_enabled`, `responsive_web_grok_share_attachment_enabled`, `responsive_web_grok_annotations_enabled`, `responsive_web_edit_tweet_api_enabled`, `rweb_conversational_replies_downvote_enabled`, `graphql_is_translatable_rweb_tweet_is_translatable_enabled`, `view_counts_everywhere_api_enabled`, `longform_notetweets_consumption_enabled`, `responsive_web_twitter_article_tweet_consumption_enabled`, `content_disclosure_indicator_enabled`, `content_disclosure_ai_generated_indicator_enabled`, `responsive_web_grok_show_grok_translated_post`, `responsive_web_grok_analysis_button_from_backend`, `post_ctas_fetch_enabled`, `longform_notetweets_rich_text_read_enabled`, `longform_notetweets_inline_media_enabled`, `profile_label_improvements_pcf_label_in_post_enabled`, `responsive_web_profile_redirect_enabled`, `rweb_tipjar_consumption_enabled`, `verified_phone_label_enabled`, `responsive_web_nested_quote_preview_enabled`, `articles_preview_enabled`, `rweb_cashtags_enabled`, `responsive_web_grok_community_note_auto_translation_is_enabled`, `freedom_of_speech_not_reach_fetch_enabled`, `standardized_nudges_misinfo`, `tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled`, `responsive_web_grok_image_annotation_enabled`, `responsive_web_grok_imagine_annotation_enabled`, `responsive_web_graphql_timeline_navigation_enabled` |
| `DataSaverMode` | 0 | — |
| `DeleteBookmark` | 0 | — |
| `DeleteContentDisclosure` | 0 | — |
| `DeleteHighlight` | 0 | — |
| `DeleteRetweet` | 0 | — |
| `DeleteTweet` | 0 | — |
| `DisableVerifiedPhoneLabel` | 0 | — |
| `dmBlockUser` | 0 | — |
| `DmNsfwMediaFilterUpdate` | 0 | — |
| `dmUnblockUser` | 0 | — |
| `DownvoteTweet` | 0 | — |
| `EnableVerifiedPhoneLabel` | 0 | — |
| `ExplorePage` | 40 | `rweb_video_screen_enabled`, `profile_label_improvements_pcf_label_in_post_enabled`, `responsive_web_profile_redirect_enabled`, `rweb_tipjar_consumption_enabled`, `verified_phone_label_enabled`, `responsive_web_graphql_timeline_navigation_enabled`, `rweb_cashtags_enabled`, `creator_subscriptions_tweet_preview_api_enabled`, `premium_content_api_read_enabled`, `communities_web_enable_tweet_community_results_fetch`, `c9s_tweet_anatomy_moderator_badge_enabled`, `responsive_web_grok_analyze_button_fetch_trends_enabled`, `responsive_web_grok_analyze_post_followups_enabled`, `rweb_cashtags_composer_attachment_enabled`, `responsive_web_jetfuel_frame`, `rweb_sports_post_context_enabled`, `responsive_web_grok_share_attachment_enabled`, `responsive_web_grok_annotations_enabled`, `articles_preview_enabled`, `responsive_web_edit_tweet_api_enabled`, `rweb_conversational_replies_downvote_enabled`, `graphql_is_translatable_rweb_tweet_is_translatable_enabled`, `view_counts_everywhere_api_enabled`, `longform_notetweets_consumption_enabled`, `responsive_web_twitter_article_tweet_consumption_enabled`, `content_disclosure_indicator_enabled`, `content_disclosure_ai_generated_indicator_enabled`, `responsive_web_grok_show_grok_translated_post`, `responsive_web_grok_analysis_button_from_backend`, `post_ctas_fetch_enabled`, `freedom_of_speech_not_reach_fetch_enabled`, `standardized_nudges_misinfo`, `tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled`, `longform_notetweets_rich_text_read_enabled`, `longform_notetweets_inline_media_enabled`, `responsive_web_nested_quote_preview_enabled`, `responsive_web_grok_image_annotation_enabled`, `responsive_web_grok_imagine_annotation_enabled`, `responsive_web_grok_community_note_auto_translation_is_enabled`, `responsive_web_enhance_cards_enabled` |
| `ExploreSidebar` | 40 | `rweb_video_screen_enabled`, `rweb_cashtags_enabled`, `profile_label_improvements_pcf_label_in_post_enabled`, `responsive_web_profile_redirect_enabled`, `rweb_tipjar_consumption_enabled`, `verified_phone_label_enabled`, `creator_subscriptions_tweet_preview_api_enabled`, `responsive_web_graphql_timeline_navigation_enabled`, `premium_content_api_read_enabled`, `communities_web_enable_tweet_community_results_fetch`, `c9s_tweet_anatomy_moderator_badge_enabled`, `responsive_web_grok_analyze_button_fetch_trends_enabled`, `responsive_web_grok_analyze_post_followups_enabled`, `rweb_cashtags_composer_attachment_enabled`, `responsive_web_jetfuel_frame`, `rweb_sports_post_context_enabled`, `responsive_web_grok_share_attachment_enabled`, `responsive_web_grok_annotations_enabled`, `articles_preview_enabled`, `responsive_web_edit_tweet_api_enabled`, `rweb_conversational_replies_downvote_enabled`, `graphql_is_translatable_rweb_tweet_is_translatable_enabled`, `view_counts_everywhere_api_enabled`, `longform_notetweets_consumption_enabled`, `responsive_web_twitter_article_tweet_consumption_enabled`, `content_disclosure_indicator_enabled`, `content_disclosure_ai_generated_indicator_enabled`, `responsive_web_grok_show_grok_translated_post`, `responsive_web_grok_analysis_button_from_backend`, `post_ctas_fetch_enabled`, `freedom_of_speech_not_reach_fetch_enabled`, `standardized_nudges_misinfo`, `tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled`, `longform_notetweets_rich_text_read_enabled`, `longform_notetweets_inline_media_enabled`, `responsive_web_nested_quote_preview_enabled`, `responsive_web_grok_image_annotation_enabled`, `responsive_web_grok_imagine_annotation_enabled`, `responsive_web_grok_community_note_auto_translation_is_enabled`, `responsive_web_enhance_cards_enabled` |
| `FavoriteTweet` | 0 | — |
| `FinanceSearchTags` | 0 | — |
| `Followers` | 40 | `rweb_video_screen_enabled`, `rweb_cashtags_enabled`, `profile_label_improvements_pcf_label_in_post_enabled`, `responsive_web_profile_redirect_enabled`, `rweb_tipjar_consumption_enabled`, `verified_phone_label_enabled`, `creator_subscriptions_tweet_preview_api_enabled`, `responsive_web_graphql_timeline_navigation_enabled`, `premium_content_api_read_enabled`, `communities_web_enable_tweet_community_results_fetch`, `c9s_tweet_anatomy_moderator_badge_enabled`, `responsive_web_grok_analyze_button_fetch_trends_enabled`, `responsive_web_grok_analyze_post_followups_enabled`, `rweb_cashtags_composer_attachment_enabled`, `responsive_web_jetfuel_frame`, `rweb_sports_post_context_enabled`, `responsive_web_grok_share_attachment_enabled`, `responsive_web_grok_annotations_enabled`, `articles_preview_enabled`, `responsive_web_edit_tweet_api_enabled`, `rweb_conversational_replies_downvote_enabled`, `graphql_is_translatable_rweb_tweet_is_translatable_enabled`, `view_counts_everywhere_api_enabled`, `longform_notetweets_consumption_enabled`, `responsive_web_twitter_article_tweet_consumption_enabled`, `content_disclosure_indicator_enabled`, `content_disclosure_ai_generated_indicator_enabled`, `responsive_web_grok_show_grok_translated_post`, `responsive_web_grok_analysis_button_from_backend`, `post_ctas_fetch_enabled`, `freedom_of_speech_not_reach_fetch_enabled`, `standardized_nudges_misinfo`, `tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled`, `longform_notetweets_rich_text_read_enabled`, `longform_notetweets_inline_media_enabled`, `responsive_web_nested_quote_preview_enabled`, `responsive_web_grok_image_annotation_enabled`, `responsive_web_grok_imagine_annotation_enabled`, `responsive_web_grok_community_note_auto_translation_is_enabled`, `responsive_web_enhance_cards_enabled` |
| `FollowersYouKnow` | 40 | `rweb_video_screen_enabled`, `rweb_cashtags_enabled`, `profile_label_improvements_pcf_label_in_post_enabled`, `responsive_web_profile_redirect_enabled`, `rweb_tipjar_consumption_enabled`, `verified_phone_label_enabled`, `creator_subscriptions_tweet_preview_api_enabled`, `responsive_web_graphql_timeline_navigation_enabled`, `premium_content_api_read_enabled`, `communities_web_enable_tweet_community_results_fetch`, `c9s_tweet_anatomy_moderator_badge_enabled`, `responsive_web_grok_analyze_button_fetch_trends_enabled`, `responsive_web_grok_analyze_post_followups_enabled`, `rweb_cashtags_composer_attachment_enabled`, `responsive_web_jetfuel_frame`, `rweb_sports_post_context_enabled`, `responsive_web_grok_share_attachment_enabled`, `responsive_web_grok_annotations_enabled`, `articles_preview_enabled`, `responsive_web_edit_tweet_api_enabled`, `rweb_conversational_replies_downvote_enabled`, `graphql_is_translatable_rweb_tweet_is_translatable_enabled`, `view_counts_everywhere_api_enabled`, `longform_notetweets_consumption_enabled`, `responsive_web_twitter_article_tweet_consumption_enabled`, `content_disclosure_indicator_enabled`, `content_disclosure_ai_generated_indicator_enabled`, `responsive_web_grok_show_grok_translated_post`, `responsive_web_grok_analysis_button_from_backend`, `post_ctas_fetch_enabled`, `freedom_of_speech_not_reach_fetch_enabled`, `standardized_nudges_misinfo`, `tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled`, `longform_notetweets_rich_text_read_enabled`, `longform_notetweets_inline_media_enabled`, `responsive_web_nested_quote_preview_enabled`, `responsive_web_grok_image_annotation_enabled`, `responsive_web_grok_imagine_annotation_enabled`, `responsive_web_grok_community_note_auto_translation_is_enabled`, `responsive_web_enhance_cards_enabled` |
| `Following` | 40 | `rweb_video_screen_enabled`, `rweb_cashtags_enabled`, `profile_label_improvements_pcf_label_in_post_enabled`, `responsive_web_profile_redirect_enabled`, `rweb_tipjar_consumption_enabled`, `verified_phone_label_enabled`, `creator_subscriptions_tweet_preview_api_enabled`, `responsive_web_graphql_timeline_navigation_enabled`, `premium_content_api_read_enabled`, `communities_web_enable_tweet_community_results_fetch`, `c9s_tweet_anatomy_moderator_badge_enabled`, `responsive_web_grok_analyze_button_fetch_trends_enabled`, `responsive_web_grok_analyze_post_followups_enabled`, `rweb_cashtags_composer_attachment_enabled`, `responsive_web_jetfuel_frame`, `rweb_sports_post_context_enabled`, `responsive_web_grok_share_attachment_enabled`, `responsive_web_grok_annotations_enabled`, `articles_preview_enabled`, `responsive_web_edit_tweet_api_enabled`, `rweb_conversational_replies_downvote_enabled`, `graphql_is_translatable_rweb_tweet_is_translatable_enabled`, `view_counts_everywhere_api_enabled`, `longform_notetweets_consumption_enabled`, `responsive_web_twitter_article_tweet_consumption_enabled`, `content_disclosure_indicator_enabled`, `content_disclosure_ai_generated_indicator_enabled`, `responsive_web_grok_show_grok_translated_post`, `responsive_web_grok_analysis_button_from_backend`, `post_ctas_fetch_enabled`, `freedom_of_speech_not_reach_fetch_enabled`, `standardized_nudges_misinfo`, `tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled`, `longform_notetweets_rich_text_read_enabled`, `longform_notetweets_inline_media_enabled`, `responsive_web_nested_quote_preview_enabled`, `responsive_web_grok_image_annotation_enabled`, `responsive_web_grok_imagine_annotation_enabled`, `responsive_web_grok_community_note_auto_translation_is_enabled`, `responsive_web_enhance_cards_enabled` |
| `GenericTimelineById` | 40 | `rweb_video_screen_enabled`, `rweb_cashtags_enabled`, `profile_label_improvements_pcf_label_in_post_enabled`, `responsive_web_profile_redirect_enabled`, `rweb_tipjar_consumption_enabled`, `verified_phone_label_enabled`, `creator_subscriptions_tweet_preview_api_enabled`, `responsive_web_graphql_timeline_navigation_enabled`, `premium_content_api_read_enabled`, `communities_web_enable_tweet_community_results_fetch`, `c9s_tweet_anatomy_moderator_badge_enabled`, `responsive_web_grok_analyze_button_fetch_trends_enabled`, `responsive_web_grok_analyze_post_followups_enabled`, `rweb_cashtags_composer_attachment_enabled`, `responsive_web_jetfuel_frame`, `rweb_sports_post_context_enabled`, `responsive_web_grok_share_attachment_enabled`, `responsive_web_grok_annotations_enabled`, `articles_preview_enabled`, `responsive_web_edit_tweet_api_enabled`, `rweb_conversational_replies_downvote_enabled`, `graphql_is_translatable_rweb_tweet_is_translatable_enabled`, `view_counts_everywhere_api_enabled`, `longform_notetweets_consumption_enabled`, `responsive_web_twitter_article_tweet_consumption_enabled`, `content_disclosure_indicator_enabled`, `content_disclosure_ai_generated_indicator_enabled`, `responsive_web_grok_show_grok_translated_post`, `responsive_web_grok_analysis_button_from_backend`, `post_ctas_fetch_enabled`, `freedom_of_speech_not_reach_fetch_enabled`, `standardized_nudges_misinfo`, `tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled`, `longform_notetweets_rich_text_read_enabled`, `longform_notetweets_inline_media_enabled`, `responsive_web_nested_quote_preview_enabled`, `responsive_web_grok_image_annotation_enabled`, `responsive_web_grok_imagine_annotation_enabled`, `responsive_web_grok_community_note_auto_translation_is_enabled`, `responsive_web_enhance_cards_enabled` |
| `getAltTextPromptPreference` | 0 | — |
| `GetUserClaims` | 0 | — |
| `GetUsernameAvailabilityAndSuggestions` | 0 | — |
| `GlobalCommunitiesLatestPostSearchTimeline` | 40 | `rweb_video_screen_enabled`, `rweb_cashtags_enabled`, `profile_label_improvements_pcf_label_in_post_enabled`, `responsive_web_profile_redirect_enabled`, `rweb_tipjar_consumption_enabled`, `verified_phone_label_enabled`, `creator_subscriptions_tweet_preview_api_enabled`, `responsive_web_graphql_timeline_navigation_enabled`, `premium_content_api_read_enabled`, `communities_web_enable_tweet_community_results_fetch`, `c9s_tweet_anatomy_moderator_badge_enabled`, `responsive_web_grok_analyze_button_fetch_trends_enabled`, `responsive_web_grok_analyze_post_followups_enabled`, `rweb_cashtags_composer_attachment_enabled`, `responsive_web_jetfuel_frame`, `rweb_sports_post_context_enabled`, `responsive_web_grok_share_attachment_enabled`, `responsive_web_grok_annotations_enabled`, `articles_preview_enabled`, `responsive_web_edit_tweet_api_enabled`, `rweb_conversational_replies_downvote_enabled`, `graphql_is_translatable_rweb_tweet_is_translatable_enabled`, `view_counts_everywhere_api_enabled`, `longform_notetweets_consumption_enabled`, `responsive_web_twitter_article_tweet_consumption_enabled`, `content_disclosure_indicator_enabled`, `content_disclosure_ai_generated_indicator_enabled`, `responsive_web_grok_show_grok_translated_post`, `responsive_web_grok_analysis_button_from_backend`, `post_ctas_fetch_enabled`, `freedom_of_speech_not_reach_fetch_enabled`, `standardized_nudges_misinfo`, `tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled`, `longform_notetweets_rich_text_read_enabled`, `longform_notetweets_inline_media_enabled`, `responsive_web_nested_quote_preview_enabled`, `responsive_web_grok_image_annotation_enabled`, `responsive_web_grok_imagine_annotation_enabled`, `responsive_web_grok_community_note_auto_translation_is_enabled`, `responsive_web_enhance_cards_enabled` |
| `GlobalCommunitiesPostSearchTimeline` | 40 | `rweb_video_screen_enabled`, `rweb_cashtags_enabled`, `profile_label_improvements_pcf_label_in_post_enabled`, `responsive_web_profile_redirect_enabled`, `rweb_tipjar_consumption_enabled`, `verified_phone_label_enabled`, `creator_subscriptions_tweet_preview_api_enabled`, `responsive_web_graphql_timeline_navigation_enabled`, `premium_content_api_read_enabled`, `communities_web_enable_tweet_community_results_fetch`, `c9s_tweet_anatomy_moderator_badge_enabled`, `responsive_web_grok_analyze_button_fetch_trends_enabled`, `responsive_web_grok_analyze_post_followups_enabled`, `rweb_cashtags_composer_attachment_enabled`, `responsive_web_jetfuel_frame`, `rweb_sports_post_context_enabled`, `responsive_web_grok_share_attachment_enabled`, `responsive_web_grok_annotations_enabled`, `articles_preview_enabled`, `responsive_web_edit_tweet_api_enabled`, `rweb_conversational_replies_downvote_enabled`, `graphql_is_translatable_rweb_tweet_is_translatable_enabled`, `view_counts_everywhere_api_enabled`, `longform_notetweets_consumption_enabled`, `responsive_web_twitter_article_tweet_consumption_enabled`, `content_disclosure_indicator_enabled`, `content_disclosure_ai_generated_indicator_enabled`, `responsive_web_grok_show_grok_translated_post`, `responsive_web_grok_analysis_button_from_backend`, `post_ctas_fetch_enabled`, `freedom_of_speech_not_reach_fetch_enabled`, `standardized_nudges_misinfo`, `tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled`, `longform_notetweets_rich_text_read_enabled`, `longform_notetweets_inline_media_enabled`, `responsive_web_nested_quote_preview_enabled`, `responsive_web_grok_image_annotation_enabled`, `responsive_web_grok_imagine_annotation_enabled`, `responsive_web_grok_community_note_auto_translation_is_enabled`, `responsive_web_enhance_cards_enabled` |
| `Likes` | 40 | `rweb_video_screen_enabled`, `rweb_cashtags_enabled`, `profile_label_improvements_pcf_label_in_post_enabled`, `responsive_web_profile_redirect_enabled`, `rweb_tipjar_consumption_enabled`, `verified_phone_label_enabled`, `creator_subscriptions_tweet_preview_api_enabled`, `responsive_web_graphql_timeline_navigation_enabled`, `premium_content_api_read_enabled`, `communities_web_enable_tweet_community_results_fetch`, `c9s_tweet_anatomy_moderator_badge_enabled`, `responsive_web_grok_analyze_button_fetch_trends_enabled`, `responsive_web_grok_analyze_post_followups_enabled`, `rweb_cashtags_composer_attachment_enabled`, `responsive_web_jetfuel_frame`, `rweb_sports_post_context_enabled`, `responsive_web_grok_share_attachment_enabled`, `responsive_web_grok_annotations_enabled`, `articles_preview_enabled`, `responsive_web_edit_tweet_api_enabled`, `rweb_conversational_replies_downvote_enabled`, `graphql_is_translatable_rweb_tweet_is_translatable_enabled`, `view_counts_everywhere_api_enabled`, `longform_notetweets_consumption_enabled`, `responsive_web_twitter_article_tweet_consumption_enabled`, `content_disclosure_indicator_enabled`, `content_disclosure_ai_generated_indicator_enabled`, `responsive_web_grok_show_grok_translated_post`, `responsive_web_grok_analysis_button_from_backend`, `post_ctas_fetch_enabled`, `freedom_of_speech_not_reach_fetch_enabled`, `standardized_nudges_misinfo`, `tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled`, `longform_notetweets_rich_text_read_enabled`, `longform_notetweets_inline_media_enabled`, `responsive_web_nested_quote_preview_enabled`, `responsive_web_grok_image_annotation_enabled`, `responsive_web_grok_imagine_annotation_enabled`, `responsive_web_grok_community_note_auto_translation_is_enabled`, `responsive_web_enhance_cards_enabled` |
| `ListSearchTimeline` | 40 | `rweb_video_screen_enabled`, `rweb_cashtags_enabled`, `profile_label_improvements_pcf_label_in_post_enabled`, `responsive_web_profile_redirect_enabled`, `rweb_tipjar_consumption_enabled`, `verified_phone_label_enabled`, `creator_subscriptions_tweet_preview_api_enabled`, `responsive_web_graphql_timeline_navigation_enabled`, `premium_content_api_read_enabled`, `communities_web_enable_tweet_community_results_fetch`, `c9s_tweet_anatomy_moderator_badge_enabled`, `responsive_web_grok_analyze_button_fetch_trends_enabled`, `responsive_web_grok_analyze_post_followups_enabled`, `rweb_cashtags_composer_attachment_enabled`, `responsive_web_jetfuel_frame`, `rweb_sports_post_context_enabled`, `responsive_web_grok_share_attachment_enabled`, `responsive_web_grok_annotations_enabled`, `articles_preview_enabled`, `responsive_web_edit_tweet_api_enabled`, `rweb_conversational_replies_downvote_enabled`, `graphql_is_translatable_rweb_tweet_is_translatable_enabled`, `view_counts_everywhere_api_enabled`, `longform_notetweets_consumption_enabled`, `responsive_web_twitter_article_tweet_consumption_enabled`, `content_disclosure_indicator_enabled`, `content_disclosure_ai_generated_indicator_enabled`, `responsive_web_grok_show_grok_translated_post`, `responsive_web_grok_analysis_button_from_backend`, `post_ctas_fetch_enabled`, `freedom_of_speech_not_reach_fetch_enabled`, `standardized_nudges_misinfo`, `tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled`, `longform_notetweets_rich_text_read_enabled`, `longform_notetweets_inline_media_enabled`, `responsive_web_nested_quote_preview_enabled`, `responsive_web_grok_image_annotation_enabled`, `responsive_web_grok_imagine_annotation_enabled`, `responsive_web_grok_community_note_auto_translation_is_enabled`, `responsive_web_enhance_cards_enabled` |
| `MediaTabVideoMixer` | 40 | `rweb_video_screen_enabled`, `rweb_cashtags_enabled`, `profile_label_improvements_pcf_label_in_post_enabled`, `responsive_web_profile_redirect_enabled`, `rweb_tipjar_consumption_enabled`, `verified_phone_label_enabled`, `creator_subscriptions_tweet_preview_api_enabled`, `responsive_web_graphql_timeline_navigation_enabled`, `premium_content_api_read_enabled`, `communities_web_enable_tweet_community_results_fetch`, `c9s_tweet_anatomy_moderator_badge_enabled`, `responsive_web_grok_analyze_button_fetch_trends_enabled`, `responsive_web_grok_analyze_post_followups_enabled`, `rweb_cashtags_composer_attachment_enabled`, `responsive_web_jetfuel_frame`, `rweb_sports_post_context_enabled`, `responsive_web_grok_share_attachment_enabled`, `responsive_web_grok_annotations_enabled`, `articles_preview_enabled`, `responsive_web_edit_tweet_api_enabled`, `rweb_conversational_replies_downvote_enabled`, `graphql_is_translatable_rweb_tweet_is_translatable_enabled`, `view_counts_everywhere_api_enabled`, `longform_notetweets_consumption_enabled`, `responsive_web_twitter_article_tweet_consumption_enabled`, `content_disclosure_indicator_enabled`, `content_disclosure_ai_generated_indicator_enabled`, `responsive_web_grok_show_grok_translated_post`, `responsive_web_grok_analysis_button_from_backend`, `post_ctas_fetch_enabled`, `freedom_of_speech_not_reach_fetch_enabled`, `standardized_nudges_misinfo`, `tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled`, `longform_notetweets_rich_text_read_enabled`, `longform_notetweets_inline_media_enabled`, `responsive_web_nested_quote_preview_enabled`, `responsive_web_grok_image_annotation_enabled`, `responsive_web_grok_imagine_annotation_enabled`, `responsive_web_grok_community_note_auto_translation_is_enabled`, `responsive_web_enhance_cards_enabled` |
| `ModeratedTimeline` | 40 | `rweb_video_screen_enabled`, `rweb_cashtags_enabled`, `profile_label_improvements_pcf_label_in_post_enabled`, `responsive_web_profile_redirect_enabled`, `rweb_tipjar_consumption_enabled`, `verified_phone_label_enabled`, `creator_subscriptions_tweet_preview_api_enabled`, `responsive_web_graphql_timeline_navigation_enabled`, `premium_content_api_read_enabled`, `communities_web_enable_tweet_community_results_fetch`, `c9s_tweet_anatomy_moderator_badge_enabled`, `responsive_web_grok_analyze_button_fetch_trends_enabled`, `responsive_web_grok_analyze_post_followups_enabled`, `rweb_cashtags_composer_attachment_enabled`, `responsive_web_jetfuel_frame`, `rweb_sports_post_context_enabled`, `responsive_web_grok_share_attachment_enabled`, `responsive_web_grok_annotations_enabled`, `articles_preview_enabled`, `responsive_web_edit_tweet_api_enabled`, `rweb_conversational_replies_downvote_enabled`, `graphql_is_translatable_rweb_tweet_is_translatable_enabled`, `view_counts_everywhere_api_enabled`, `longform_notetweets_consumption_enabled`, `responsive_web_twitter_article_tweet_consumption_enabled`, `content_disclosure_indicator_enabled`, `content_disclosure_ai_generated_indicator_enabled`, `responsive_web_grok_show_grok_translated_post`, `responsive_web_grok_analysis_button_from_backend`, `post_ctas_fetch_enabled`, `freedom_of_speech_not_reach_fetch_enabled`, `standardized_nudges_misinfo`, `tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled`, `longform_notetweets_rich_text_read_enabled`, `longform_notetweets_inline_media_enabled`, `responsive_web_nested_quote_preview_enabled`, `responsive_web_grok_image_annotation_enabled`, `responsive_web_grok_imagine_annotation_enabled`, `responsive_web_grok_community_note_auto_translation_is_enabled`, `responsive_web_enhance_cards_enabled` |
| `ModerateTweet` | 0 | — |
| `MutedAccounts` | 40 | `rweb_video_screen_enabled`, `rweb_cashtags_enabled`, `profile_label_improvements_pcf_label_in_post_enabled`, `responsive_web_profile_redirect_enabled`, `rweb_tipjar_consumption_enabled`, `verified_phone_label_enabled`, `creator_subscriptions_tweet_preview_api_enabled`, `responsive_web_graphql_timeline_navigation_enabled`, `premium_content_api_read_enabled`, `communities_web_enable_tweet_community_results_fetch`, `c9s_tweet_anatomy_moderator_badge_enabled`, `responsive_web_grok_analyze_button_fetch_trends_enabled`, `responsive_web_grok_analyze_post_followups_enabled`, `rweb_cashtags_composer_attachment_enabled`, `responsive_web_jetfuel_frame`, `rweb_sports_post_context_enabled`, `responsive_web_grok_share_attachment_enabled`, `responsive_web_grok_annotations_enabled`, `articles_preview_enabled`, `responsive_web_edit_tweet_api_enabled`, `rweb_conversational_replies_downvote_enabled`, `graphql_is_translatable_rweb_tweet_is_translatable_enabled`, `view_counts_everywhere_api_enabled`, `longform_notetweets_consumption_enabled`, `responsive_web_twitter_article_tweet_consumption_enabled`, `content_disclosure_indicator_enabled`, `content_disclosure_ai_generated_indicator_enabled`, `responsive_web_grok_show_grok_translated_post`, `responsive_web_grok_analysis_button_from_backend`, `post_ctas_fetch_enabled`, `freedom_of_speech_not_reach_fetch_enabled`, `standardized_nudges_misinfo`, `tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled`, `longform_notetweets_rich_text_read_enabled`, `longform_notetweets_inline_media_enabled`, `responsive_web_nested_quote_preview_enabled`, `responsive_web_grok_image_annotation_enabled`, `responsive_web_grok_imagine_annotation_enabled`, `responsive_web_grok_community_note_auto_translation_is_enabled`, `responsive_web_enhance_cards_enabled` |
| `PaymentsUsersTypeahead` | 1 | `responsive_web_graphql_timeline_navigation_enabled` |
| `PinnableTimelines` | 5 | `profile_label_improvements_pcf_label_in_post_enabled`, `responsive_web_profile_redirect_enabled`, `rweb_tipjar_consumption_enabled`, `verified_phone_label_enabled`, `responsive_web_graphql_timeline_navigation_enabled` |
| `PinnedTimelines` | 5 | `profile_label_improvements_pcf_label_in_post_enabled`, `responsive_web_profile_redirect_enabled`, `rweb_tipjar_consumption_enabled`, `verified_phone_label_enabled`, `responsive_web_graphql_timeline_navigation_enabled` |
| `PinTimeline` | 5 | `profile_label_improvements_pcf_label_in_post_enabled`, `responsive_web_profile_redirect_enabled`, `rweb_tipjar_consumption_enabled`, `verified_phone_label_enabled`, `responsive_web_graphql_timeline_navigation_enabled` |
| `PinTweet` | 0 | — |
| `ProfileFilter` | 40 | `rweb_video_screen_enabled`, `rweb_cashtags_enabled`, `profile_label_improvements_pcf_label_in_post_enabled`, `responsive_web_profile_redirect_enabled`, `rweb_tipjar_consumption_enabled`, `verified_phone_label_enabled`, `creator_subscriptions_tweet_preview_api_enabled`, `responsive_web_graphql_timeline_navigation_enabled`, `premium_content_api_read_enabled`, `communities_web_enable_tweet_community_results_fetch`, `c9s_tweet_anatomy_moderator_badge_enabled`, `responsive_web_grok_analyze_button_fetch_trends_enabled`, `responsive_web_grok_analyze_post_followups_enabled`, `rweb_cashtags_composer_attachment_enabled`, `responsive_web_jetfuel_frame`, `rweb_sports_post_context_enabled`, `responsive_web_grok_share_attachment_enabled`, `responsive_web_grok_annotations_enabled`, `articles_preview_enabled`, `responsive_web_edit_tweet_api_enabled`, `rweb_conversational_replies_downvote_enabled`, `graphql_is_translatable_rweb_tweet_is_translatable_enabled`, `view_counts_everywhere_api_enabled`, `longform_notetweets_consumption_enabled`, `responsive_web_twitter_article_tweet_consumption_enabled`, `content_disclosure_indicator_enabled`, `content_disclosure_ai_generated_indicator_enabled`, `responsive_web_grok_show_grok_translated_post`, `responsive_web_grok_analysis_button_from_backend`, `post_ctas_fetch_enabled`, `freedom_of_speech_not_reach_fetch_enabled`, `standardized_nudges_misinfo`, `tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled`, `longform_notetweets_rich_text_read_enabled`, `longform_notetweets_inline_media_enabled`, `responsive_web_nested_quote_preview_enabled`, `responsive_web_grok_image_annotation_enabled`, `responsive_web_grok_imagine_annotation_enabled`, `responsive_web_grok_community_note_auto_translation_is_enabled`, `responsive_web_enhance_cards_enabled` |
| `ProfileUserPhoneState` | 0 | — |
| `PutClientEducationFlag` | 0 | — |
| `RemoveFollower` | 0 | — |
| `SearchTimeline` | 40 | `rweb_video_screen_enabled`, `rweb_cashtags_enabled`, `profile_label_improvements_pcf_label_in_post_enabled`, `responsive_web_profile_redirect_enabled`, `rweb_tipjar_consumption_enabled`, `verified_phone_label_enabled`, `creator_subscriptions_tweet_preview_api_enabled`, `responsive_web_graphql_timeline_navigation_enabled`, `premium_content_api_read_enabled`, `communities_web_enable_tweet_community_results_fetch`, `c9s_tweet_anatomy_moderator_badge_enabled`, `responsive_web_grok_analyze_button_fetch_trends_enabled`, `responsive_web_grok_analyze_post_followups_enabled`, `rweb_cashtags_composer_attachment_enabled`, `responsive_web_jetfuel_frame`, `rweb_sports_post_context_enabled`, `responsive_web_grok_share_attachment_enabled`, `responsive_web_grok_annotations_enabled`, `articles_preview_enabled`, `responsive_web_edit_tweet_api_enabled`, `rweb_conversational_replies_downvote_enabled`, `graphql_is_translatable_rweb_tweet_is_translatable_enabled`, `view_counts_everywhere_api_enabled`, `longform_notetweets_consumption_enabled`, `responsive_web_twitter_article_tweet_consumption_enabled`, `content_disclosure_indicator_enabled`, `content_disclosure_ai_generated_indicator_enabled`, `responsive_web_grok_show_grok_translated_post`, `responsive_web_grok_analysis_button_from_backend`, `post_ctas_fetch_enabled`, `freedom_of_speech_not_reach_fetch_enabled`, `standardized_nudges_misinfo`, `tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled`, `longform_notetweets_rich_text_read_enabled`, `longform_notetweets_inline_media_enabled`, `responsive_web_nested_quote_preview_enabled`, `responsive_web_grok_image_annotation_enabled`, `responsive_web_grok_imagine_annotation_enabled`, `responsive_web_grok_community_note_auto_translation_is_enabled`, `responsive_web_enhance_cards_enabled` |
| `SharingAudiospacesListeningDataWithFollowersUpdate` | 0 | — |
| `SimilarPosts` | 40 | `rweb_video_screen_enabled`, `rweb_cashtags_enabled`, `profile_label_improvements_pcf_label_in_post_enabled`, `responsive_web_profile_redirect_enabled`, `rweb_tipjar_consumption_enabled`, `verified_phone_label_enabled`, `creator_subscriptions_tweet_preview_api_enabled`, `responsive_web_graphql_timeline_navigation_enabled`, `premium_content_api_read_enabled`, `communities_web_enable_tweet_community_results_fetch`, `c9s_tweet_anatomy_moderator_badge_enabled`, `responsive_web_grok_analyze_button_fetch_trends_enabled`, `responsive_web_grok_analyze_post_followups_enabled`, `rweb_cashtags_composer_attachment_enabled`, `responsive_web_jetfuel_frame`, `rweb_sports_post_context_enabled`, `responsive_web_grok_share_attachment_enabled`, `responsive_web_grok_annotations_enabled`, `articles_preview_enabled`, `responsive_web_edit_tweet_api_enabled`, `rweb_conversational_replies_downvote_enabled`, `graphql_is_translatable_rweb_tweet_is_translatable_enabled`, `view_counts_everywhere_api_enabled`, `longform_notetweets_consumption_enabled`, `responsive_web_twitter_article_tweet_consumption_enabled`, `content_disclosure_indicator_enabled`, `content_disclosure_ai_generated_indicator_enabled`, `responsive_web_grok_show_grok_translated_post`, `responsive_web_grok_analysis_button_from_backend`, `post_ctas_fetch_enabled`, `freedom_of_speech_not_reach_fetch_enabled`, `standardized_nudges_misinfo`, `tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled`, `longform_notetweets_rich_text_read_enabled`, `longform_notetweets_inline_media_enabled`, `responsive_web_nested_quote_preview_enabled`, `responsive_web_grok_image_annotation_enabled`, `responsive_web_grok_imagine_annotation_enabled`, `responsive_web_grok_community_note_auto_translation_is_enabled`, `responsive_web_enhance_cards_enabled` |
| `SuperFollowers` | 40 | `rweb_video_screen_enabled`, `rweb_cashtags_enabled`, `profile_label_improvements_pcf_label_in_post_enabled`, `responsive_web_profile_redirect_enabled`, `rweb_tipjar_consumption_enabled`, `verified_phone_label_enabled`, `creator_subscriptions_tweet_preview_api_enabled`, `responsive_web_graphql_timeline_navigation_enabled`, `premium_content_api_read_enabled`, `communities_web_enable_tweet_community_results_fetch`, `c9s_tweet_anatomy_moderator_badge_enabled`, `responsive_web_grok_analyze_button_fetch_trends_enabled`, `responsive_web_grok_analyze_post_followups_enabled`, `rweb_cashtags_composer_attachment_enabled`, `responsive_web_jetfuel_frame`, `rweb_sports_post_context_enabled`, `responsive_web_grok_share_attachment_enabled`, `responsive_web_grok_annotations_enabled`, `articles_preview_enabled`, `responsive_web_edit_tweet_api_enabled`, `rweb_conversational_replies_downvote_enabled`, `graphql_is_translatable_rweb_tweet_is_translatable_enabled`, `view_counts_everywhere_api_enabled`, `longform_notetweets_consumption_enabled`, `responsive_web_twitter_article_tweet_consumption_enabled`, `content_disclosure_indicator_enabled`, `content_disclosure_ai_generated_indicator_enabled`, `responsive_web_grok_show_grok_translated_post`, `responsive_web_grok_analysis_button_from_backend`, `post_ctas_fetch_enabled`, `freedom_of_speech_not_reach_fetch_enabled`, `standardized_nudges_misinfo`, `tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled`, `longform_notetweets_rich_text_read_enabled`, `longform_notetweets_inline_media_enabled`, `responsive_web_nested_quote_preview_enabled`, `responsive_web_grok_image_annotation_enabled`, `responsive_web_grok_imagine_annotation_enabled`, `responsive_web_grok_community_note_auto_translation_is_enabled`, `responsive_web_enhance_cards_enabled` |
| `SupportedLanguages` | 0 | — |
| `timelinesFeedback` | 0 | — |
| `TrendHistory` | 40 | `rweb_video_screen_enabled`, `rweb_cashtags_enabled`, `profile_label_improvements_pcf_label_in_post_enabled`, `responsive_web_profile_redirect_enabled`, `rweb_tipjar_consumption_enabled`, `verified_phone_label_enabled`, `creator_subscriptions_tweet_preview_api_enabled`, `responsive_web_graphql_timeline_navigation_enabled`, `premium_content_api_read_enabled`, `communities_web_enable_tweet_community_results_fetch`, `c9s_tweet_anatomy_moderator_badge_enabled`, `responsive_web_grok_analyze_button_fetch_trends_enabled`, `responsive_web_grok_analyze_post_followups_enabled`, `rweb_cashtags_composer_attachment_enabled`, `responsive_web_jetfuel_frame`, `rweb_sports_post_context_enabled`, `responsive_web_grok_share_attachment_enabled`, `responsive_web_grok_annotations_enabled`, `articles_preview_enabled`, `responsive_web_edit_tweet_api_enabled`, `rweb_conversational_replies_downvote_enabled`, `graphql_is_translatable_rweb_tweet_is_translatable_enabled`, `view_counts_everywhere_api_enabled`, `longform_notetweets_consumption_enabled`, `responsive_web_twitter_article_tweet_consumption_enabled`, `content_disclosure_indicator_enabled`, `content_disclosure_ai_generated_indicator_enabled`, `responsive_web_grok_show_grok_translated_post`, `responsive_web_grok_analysis_button_from_backend`, `post_ctas_fetch_enabled`, `freedom_of_speech_not_reach_fetch_enabled`, `standardized_nudges_misinfo`, `tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled`, `longform_notetweets_rich_text_read_enabled`, `longform_notetweets_inline_media_enabled`, `responsive_web_nested_quote_preview_enabled`, `responsive_web_grok_image_annotation_enabled`, `responsive_web_grok_imagine_annotation_enabled`, `responsive_web_grok_community_note_auto_translation_is_enabled`, `responsive_web_enhance_cards_enabled` |
| `TrendRelevantUsers` | 40 | `rweb_video_screen_enabled`, `rweb_cashtags_enabled`, `profile_label_improvements_pcf_label_in_post_enabled`, `responsive_web_profile_redirect_enabled`, `rweb_tipjar_consumption_enabled`, `verified_phone_label_enabled`, `creator_subscriptions_tweet_preview_api_enabled`, `responsive_web_graphql_timeline_navigation_enabled`, `premium_content_api_read_enabled`, `communities_web_enable_tweet_community_results_fetch`, `c9s_tweet_anatomy_moderator_badge_enabled`, `responsive_web_grok_analyze_button_fetch_trends_enabled`, `responsive_web_grok_analyze_post_followups_enabled`, `rweb_cashtags_composer_attachment_enabled`, `responsive_web_jetfuel_frame`, `rweb_sports_post_context_enabled`, `responsive_web_grok_share_attachment_enabled`, `responsive_web_grok_annotations_enabled`, `articles_preview_enabled`, `responsive_web_edit_tweet_api_enabled`, `rweb_conversational_replies_downvote_enabled`, `graphql_is_translatable_rweb_tweet_is_translatable_enabled`, `view_counts_everywhere_api_enabled`, `longform_notetweets_consumption_enabled`, `responsive_web_twitter_article_tweet_consumption_enabled`, `content_disclosure_indicator_enabled`, `content_disclosure_ai_generated_indicator_enabled`, `responsive_web_grok_show_grok_translated_post`, `responsive_web_grok_analysis_button_from_backend`, `post_ctas_fetch_enabled`, `freedom_of_speech_not_reach_fetch_enabled`, `standardized_nudges_misinfo`, `tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled`, `longform_notetweets_rich_text_read_enabled`, `longform_notetweets_inline_media_enabled`, `responsive_web_nested_quote_preview_enabled`, `responsive_web_grok_image_annotation_enabled`, `responsive_web_grok_imagine_annotation_enabled`, `responsive_web_grok_community_note_auto_translation_is_enabled`, `responsive_web_enhance_cards_enabled` |
| `TVHomeMixer` | 40 | `rweb_video_screen_enabled`, `rweb_cashtags_enabled`, `profile_label_improvements_pcf_label_in_post_enabled`, `responsive_web_profile_redirect_enabled`, `rweb_tipjar_consumption_enabled`, `verified_phone_label_enabled`, `creator_subscriptions_tweet_preview_api_enabled`, `responsive_web_graphql_timeline_navigation_enabled`, `premium_content_api_read_enabled`, `communities_web_enable_tweet_community_results_fetch`, `c9s_tweet_anatomy_moderator_badge_enabled`, `responsive_web_grok_analyze_button_fetch_trends_enabled`, `responsive_web_grok_analyze_post_followups_enabled`, `rweb_cashtags_composer_attachment_enabled`, `responsive_web_jetfuel_frame`, `rweb_sports_post_context_enabled`, `responsive_web_grok_share_attachment_enabled`, `responsive_web_grok_annotations_enabled`, `articles_preview_enabled`, `responsive_web_edit_tweet_api_enabled`, `rweb_conversational_replies_downvote_enabled`, `graphql_is_translatable_rweb_tweet_is_translatable_enabled`, `view_counts_everywhere_api_enabled`, `longform_notetweets_consumption_enabled`, `responsive_web_twitter_article_tweet_consumption_enabled`, `content_disclosure_indicator_enabled`, `content_disclosure_ai_generated_indicator_enabled`, `responsive_web_grok_show_grok_translated_post`, `responsive_web_grok_analysis_button_from_backend`, `post_ctas_fetch_enabled`, `freedom_of_speech_not_reach_fetch_enabled`, `standardized_nudges_misinfo`, `tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled`, `longform_notetweets_rich_text_read_enabled`, `longform_notetweets_inline_media_enabled`, `responsive_web_nested_quote_preview_enabled`, `responsive_web_grok_image_annotation_enabled`, `responsive_web_grok_imagine_annotation_enabled`, `responsive_web_grok_community_note_auto_translation_is_enabled`, `responsive_web_enhance_cards_enabled` |
| `TweetDetail` | 40 | `rweb_video_screen_enabled`, `rweb_cashtags_enabled`, `profile_label_improvements_pcf_label_in_post_enabled`, `responsive_web_profile_redirect_enabled`, `rweb_tipjar_consumption_enabled`, `verified_phone_label_enabled`, `creator_subscriptions_tweet_preview_api_enabled`, `responsive_web_graphql_timeline_navigation_enabled`, `premium_content_api_read_enabled`, `communities_web_enable_tweet_community_results_fetch`, `c9s_tweet_anatomy_moderator_badge_enabled`, `responsive_web_grok_analyze_button_fetch_trends_enabled`, `responsive_web_grok_analyze_post_followups_enabled`, `rweb_cashtags_composer_attachment_enabled`, `responsive_web_jetfuel_frame`, `rweb_sports_post_context_enabled`, `responsive_web_grok_share_attachment_enabled`, `responsive_web_grok_annotations_enabled`, `articles_preview_enabled`, `responsive_web_edit_tweet_api_enabled`, `rweb_conversational_replies_downvote_enabled`, `graphql_is_translatable_rweb_tweet_is_translatable_enabled`, `view_counts_everywhere_api_enabled`, `longform_notetweets_consumption_enabled`, `responsive_web_twitter_article_tweet_consumption_enabled`, `content_disclosure_indicator_enabled`, `content_disclosure_ai_generated_indicator_enabled`, `responsive_web_grok_show_grok_translated_post`, `responsive_web_grok_analysis_button_from_backend`, `post_ctas_fetch_enabled`, `freedom_of_speech_not_reach_fetch_enabled`, `standardized_nudges_misinfo`, `tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled`, `longform_notetweets_rich_text_read_enabled`, `longform_notetweets_inline_media_enabled`, `responsive_web_nested_quote_preview_enabled`, `responsive_web_grok_image_annotation_enabled`, `responsive_web_grok_imagine_annotation_enabled`, `responsive_web_grok_community_note_auto_translation_is_enabled`, `responsive_web_enhance_cards_enabled` |
| `TweetResultByRestId` | 38 | `creator_subscriptions_tweet_preview_api_enabled`, `premium_content_api_read_enabled`, `communities_web_enable_tweet_community_results_fetch`, `c9s_tweet_anatomy_moderator_badge_enabled`, `responsive_web_grok_analyze_button_fetch_trends_enabled`, `responsive_web_grok_analyze_post_followups_enabled`, `rweb_cashtags_composer_attachment_enabled`, `responsive_web_jetfuel_frame`, `rweb_sports_post_context_enabled`, `responsive_web_grok_share_attachment_enabled`, `responsive_web_grok_annotations_enabled`, `articles_preview_enabled`, `responsive_web_edit_tweet_api_enabled`, `rweb_conversational_replies_downvote_enabled`, `graphql_is_translatable_rweb_tweet_is_translatable_enabled`, `view_counts_everywhere_api_enabled`, `longform_notetweets_consumption_enabled`, `responsive_web_twitter_article_tweet_consumption_enabled`, `content_disclosure_indicator_enabled`, `content_disclosure_ai_generated_indicator_enabled`, `responsive_web_grok_show_grok_translated_post`, `responsive_web_grok_analysis_button_from_backend`, `post_ctas_fetch_enabled`, `rweb_cashtags_enabled`, `freedom_of_speech_not_reach_fetch_enabled`, `standardized_nudges_misinfo`, `tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled`, `longform_notetweets_rich_text_read_enabled`, `longform_notetweets_inline_media_enabled`, `profile_label_improvements_pcf_label_in_post_enabled`, `responsive_web_profile_redirect_enabled`, `rweb_tipjar_consumption_enabled`, `verified_phone_label_enabled`, `responsive_web_nested_quote_preview_enabled`, `responsive_web_grok_image_annotation_enabled`, `responsive_web_grok_imagine_annotation_enabled`, `responsive_web_grok_community_note_auto_translation_is_enabled`, `responsive_web_graphql_timeline_navigation_enabled` |
| `TweetResultsByRestIds` | 38 | `creator_subscriptions_tweet_preview_api_enabled`, `premium_content_api_read_enabled`, `communities_web_enable_tweet_community_results_fetch`, `c9s_tweet_anatomy_moderator_badge_enabled`, `responsive_web_grok_analyze_button_fetch_trends_enabled`, `responsive_web_grok_analyze_post_followups_enabled`, `rweb_cashtags_composer_attachment_enabled`, `responsive_web_jetfuel_frame`, `rweb_sports_post_context_enabled`, `responsive_web_grok_share_attachment_enabled`, `responsive_web_grok_annotations_enabled`, `articles_preview_enabled`, `responsive_web_edit_tweet_api_enabled`, `rweb_conversational_replies_downvote_enabled`, `graphql_is_translatable_rweb_tweet_is_translatable_enabled`, `view_counts_everywhere_api_enabled`, `longform_notetweets_consumption_enabled`, `responsive_web_twitter_article_tweet_consumption_enabled`, `content_disclosure_indicator_enabled`, `content_disclosure_ai_generated_indicator_enabled`, `responsive_web_grok_show_grok_translated_post`, `responsive_web_grok_analysis_button_from_backend`, `post_ctas_fetch_enabled`, `freedom_of_speech_not_reach_fetch_enabled`, `standardized_nudges_misinfo`, `tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled`, `rweb_cashtags_enabled`, `longform_notetweets_rich_text_read_enabled`, `longform_notetweets_inline_media_enabled`, `profile_label_improvements_pcf_label_in_post_enabled`, `responsive_web_profile_redirect_enabled`, `rweb_tipjar_consumption_enabled`, `verified_phone_label_enabled`, `responsive_web_nested_quote_preview_enabled`, `responsive_web_grok_image_annotation_enabled`, `responsive_web_grok_imagine_annotation_enabled`, `responsive_web_grok_community_note_auto_translation_is_enabled`, `responsive_web_graphql_timeline_navigation_enabled` |
| `UndoDownvoteTweet` | 0 | — |
| `UnfavoriteTweet` | 0 | — |
| `UnmentionUserFromConversation` | 0 | — |
| `UnmoderateTweet` | 0 | — |
| `UnpinTimeline` | 5 | `profile_label_improvements_pcf_label_in_post_enabled`, `responsive_web_profile_redirect_enabled`, `rweb_tipjar_consumption_enabled`, `verified_phone_label_enabled`, `responsive_web_graphql_timeline_navigation_enabled` |
| `UnpinTweet` | 0 | — |
| `updateAltTextPromptPreference` | 0 | — |
| `UpdatePinnedTimelines` | 5 | `profile_label_improvements_pcf_label_in_post_enabled`, `responsive_web_profile_redirect_enabled`, `rweb_tipjar_consumption_enabled`, `verified_phone_label_enabled`, `responsive_web_graphql_timeline_navigation_enabled` |
| `Upsells` | 1 | `subscriptions_upsells_api_enabled` |
| `UrtFixtures` | 40 | `rweb_video_screen_enabled`, `rweb_cashtags_enabled`, `profile_label_improvements_pcf_label_in_post_enabled`, `responsive_web_profile_redirect_enabled`, `rweb_tipjar_consumption_enabled`, `verified_phone_label_enabled`, `creator_subscriptions_tweet_preview_api_enabled`, `responsive_web_graphql_timeline_navigation_enabled`, `premium_content_api_read_enabled`, `communities_web_enable_tweet_community_results_fetch`, `c9s_tweet_anatomy_moderator_badge_enabled`, `responsive_web_grok_analyze_button_fetch_trends_enabled`, `responsive_web_grok_analyze_post_followups_enabled`, `rweb_cashtags_composer_attachment_enabled`, `responsive_web_jetfuel_frame`, `rweb_sports_post_context_enabled`, `responsive_web_grok_share_attachment_enabled`, `responsive_web_grok_annotations_enabled`, `articles_preview_enabled`, `responsive_web_edit_tweet_api_enabled`, `rweb_conversational_replies_downvote_enabled`, `graphql_is_translatable_rweb_tweet_is_translatable_enabled`, `view_counts_everywhere_api_enabled`, `longform_notetweets_consumption_enabled`, `responsive_web_twitter_article_tweet_consumption_enabled`, `content_disclosure_indicator_enabled`, `content_disclosure_ai_generated_indicator_enabled`, `responsive_web_grok_show_grok_translated_post`, `responsive_web_grok_analysis_button_from_backend`, `post_ctas_fetch_enabled`, `freedom_of_speech_not_reach_fetch_enabled`, `standardized_nudges_misinfo`, `tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled`, `longform_notetweets_rich_text_read_enabled`, `longform_notetweets_inline_media_enabled`, `responsive_web_nested_quote_preview_enabled`, `responsive_web_grok_image_annotation_enabled`, `responsive_web_grok_imagine_annotation_enabled`, `responsive_web_grok_community_note_auto_translation_is_enabled`, `responsive_web_enhance_cards_enabled` |
| `UserArticlesTweets` | 40 | `rweb_video_screen_enabled`, `rweb_cashtags_enabled`, `profile_label_improvements_pcf_label_in_post_enabled`, `responsive_web_profile_redirect_enabled`, `rweb_tipjar_consumption_enabled`, `verified_phone_label_enabled`, `creator_subscriptions_tweet_preview_api_enabled`, `responsive_web_graphql_timeline_navigation_enabled`, `premium_content_api_read_enabled`, `communities_web_enable_tweet_community_results_fetch`, `c9s_tweet_anatomy_moderator_badge_enabled`, `responsive_web_grok_analyze_button_fetch_trends_enabled`, `responsive_web_grok_analyze_post_followups_enabled`, `rweb_cashtags_composer_attachment_enabled`, `responsive_web_jetfuel_frame`, `rweb_sports_post_context_enabled`, `responsive_web_grok_share_attachment_enabled`, `responsive_web_grok_annotations_enabled`, `articles_preview_enabled`, `responsive_web_edit_tweet_api_enabled`, `rweb_conversational_replies_downvote_enabled`, `graphql_is_translatable_rweb_tweet_is_translatable_enabled`, `view_counts_everywhere_api_enabled`, `longform_notetweets_consumption_enabled`, `responsive_web_twitter_article_tweet_consumption_enabled`, `content_disclosure_indicator_enabled`, `content_disclosure_ai_generated_indicator_enabled`, `responsive_web_grok_show_grok_translated_post`, `responsive_web_grok_analysis_button_from_backend`, `post_ctas_fetch_enabled`, `freedom_of_speech_not_reach_fetch_enabled`, `standardized_nudges_misinfo`, `tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled`, `longform_notetweets_rich_text_read_enabled`, `longform_notetweets_inline_media_enabled`, `responsive_web_nested_quote_preview_enabled`, `responsive_web_grok_image_annotation_enabled`, `responsive_web_grok_imagine_annotation_enabled`, `responsive_web_grok_community_note_auto_translation_is_enabled`, `responsive_web_enhance_cards_enabled` |
| `UserBusinessProfileTeamTimeline` | 40 | `rweb_video_screen_enabled`, `rweb_cashtags_enabled`, `profile_label_improvements_pcf_label_in_post_enabled`, `responsive_web_profile_redirect_enabled`, `rweb_tipjar_consumption_enabled`, `verified_phone_label_enabled`, `creator_subscriptions_tweet_preview_api_enabled`, `responsive_web_graphql_timeline_navigation_enabled`, `premium_content_api_read_enabled`, `communities_web_enable_tweet_community_results_fetch`, `c9s_tweet_anatomy_moderator_badge_enabled`, `responsive_web_grok_analyze_button_fetch_trends_enabled`, `responsive_web_grok_analyze_post_followups_enabled`, `rweb_cashtags_composer_attachment_enabled`, `responsive_web_jetfuel_frame`, `rweb_sports_post_context_enabled`, `responsive_web_grok_share_attachment_enabled`, `responsive_web_grok_annotations_enabled`, `articles_preview_enabled`, `responsive_web_edit_tweet_api_enabled`, `rweb_conversational_replies_downvote_enabled`, `graphql_is_translatable_rweb_tweet_is_translatable_enabled`, `view_counts_everywhere_api_enabled`, `longform_notetweets_consumption_enabled`, `responsive_web_twitter_article_tweet_consumption_enabled`, `content_disclosure_indicator_enabled`, `content_disclosure_ai_generated_indicator_enabled`, `responsive_web_grok_show_grok_translated_post`, `responsive_web_grok_analysis_button_from_backend`, `post_ctas_fetch_enabled`, `freedom_of_speech_not_reach_fetch_enabled`, `standardized_nudges_misinfo`, `tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled`, `longform_notetweets_rich_text_read_enabled`, `longform_notetweets_inline_media_enabled`, `responsive_web_nested_quote_preview_enabled`, `responsive_web_grok_image_annotation_enabled`, `responsive_web_grok_imagine_annotation_enabled`, `responsive_web_grok_community_note_auto_translation_is_enabled`, `responsive_web_enhance_cards_enabled` |
| `UserByRestId` | 10 | `hidden_profile_subscriptions_enabled`, `profile_label_improvements_pcf_label_in_post_enabled`, `responsive_web_profile_redirect_enabled`, `rweb_tipjar_consumption_enabled`, `verified_phone_label_enabled`, `highlights_tweets_tab_ui_enabled`, `responsive_web_twitter_article_notes_tab_enabled`, `subscriptions_feature_can_gift_premium`, `creator_subscriptions_tweet_preview_api_enabled`, `responsive_web_graphql_timeline_navigation_enabled` |
| `UserByScreenName` | 12 | `hidden_profile_subscriptions_enabled`, `profile_label_improvements_pcf_label_in_post_enabled`, `responsive_web_profile_redirect_enabled`, `rweb_tipjar_consumption_enabled`, `verified_phone_label_enabled`, `subscriptions_verification_info_is_identity_verified_enabled`, `subscriptions_verification_info_verified_since_enabled`, `highlights_tweets_tab_ui_enabled`, `responsive_web_twitter_article_notes_tab_enabled`, `subscriptions_feature_can_gift_premium`, `creator_subscriptions_tweet_preview_api_enabled`, `responsive_web_graphql_timeline_navigation_enabled` |
| `UserCreatorSubscribers` | 40 | `rweb_video_screen_enabled`, `rweb_cashtags_enabled`, `profile_label_improvements_pcf_label_in_post_enabled`, `responsive_web_profile_redirect_enabled`, `rweb_tipjar_consumption_enabled`, `verified_phone_label_enabled`, `creator_subscriptions_tweet_preview_api_enabled`, `responsive_web_graphql_timeline_navigation_enabled`, `premium_content_api_read_enabled`, `communities_web_enable_tweet_community_results_fetch`, `c9s_tweet_anatomy_moderator_badge_enabled`, `responsive_web_grok_analyze_button_fetch_trends_enabled`, `responsive_web_grok_analyze_post_followups_enabled`, `rweb_cashtags_composer_attachment_enabled`, `responsive_web_jetfuel_frame`, `rweb_sports_post_context_enabled`, `responsive_web_grok_share_attachment_enabled`, `responsive_web_grok_annotations_enabled`, `articles_preview_enabled`, `responsive_web_edit_tweet_api_enabled`, `rweb_conversational_replies_downvote_enabled`, `graphql_is_translatable_rweb_tweet_is_translatable_enabled`, `view_counts_everywhere_api_enabled`, `longform_notetweets_consumption_enabled`, `responsive_web_twitter_article_tweet_consumption_enabled`, `content_disclosure_indicator_enabled`, `content_disclosure_ai_generated_indicator_enabled`, `responsive_web_grok_show_grok_translated_post`, `responsive_web_grok_analysis_button_from_backend`, `post_ctas_fetch_enabled`, `freedom_of_speech_not_reach_fetch_enabled`, `standardized_nudges_misinfo`, `tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled`, `longform_notetweets_rich_text_read_enabled`, `longform_notetweets_inline_media_enabled`, `responsive_web_nested_quote_preview_enabled`, `responsive_web_grok_image_annotation_enabled`, `responsive_web_grok_imagine_annotation_enabled`, `responsive_web_grok_community_note_auto_translation_is_enabled`, `responsive_web_enhance_cards_enabled` |
| `UserCreatorSubscriptions` | 40 | `rweb_video_screen_enabled`, `rweb_cashtags_enabled`, `profile_label_improvements_pcf_label_in_post_enabled`, `responsive_web_profile_redirect_enabled`, `rweb_tipjar_consumption_enabled`, `verified_phone_label_enabled`, `creator_subscriptions_tweet_preview_api_enabled`, `responsive_web_graphql_timeline_navigation_enabled`, `premium_content_api_read_enabled`, `communities_web_enable_tweet_community_results_fetch`, `c9s_tweet_anatomy_moderator_badge_enabled`, `responsive_web_grok_analyze_button_fetch_trends_enabled`, `responsive_web_grok_analyze_post_followups_enabled`, `rweb_cashtags_composer_attachment_enabled`, `responsive_web_jetfuel_frame`, `rweb_sports_post_context_enabled`, `responsive_web_grok_share_attachment_enabled`, `responsive_web_grok_annotations_enabled`, `articles_preview_enabled`, `responsive_web_edit_tweet_api_enabled`, `rweb_conversational_replies_downvote_enabled`, `graphql_is_translatable_rweb_tweet_is_translatable_enabled`, `view_counts_everywhere_api_enabled`, `longform_notetweets_consumption_enabled`, `responsive_web_twitter_article_tweet_consumption_enabled`, `content_disclosure_indicator_enabled`, `content_disclosure_ai_generated_indicator_enabled`, `responsive_web_grok_show_grok_translated_post`, `responsive_web_grok_analysis_button_from_backend`, `post_ctas_fetch_enabled`, `freedom_of_speech_not_reach_fetch_enabled`, `standardized_nudges_misinfo`, `tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled`, `longform_notetweets_rich_text_read_enabled`, `longform_notetweets_inline_media_enabled`, `responsive_web_nested_quote_preview_enabled`, `responsive_web_grok_image_annotation_enabled`, `responsive_web_grok_imagine_annotation_enabled`, `responsive_web_grok_community_note_auto_translation_is_enabled`, `responsive_web_enhance_cards_enabled` |
| `UserHighlightsTweets` | 40 | `rweb_video_screen_enabled`, `rweb_cashtags_enabled`, `profile_label_improvements_pcf_label_in_post_enabled`, `responsive_web_profile_redirect_enabled`, `rweb_tipjar_consumption_enabled`, `verified_phone_label_enabled`, `creator_subscriptions_tweet_preview_api_enabled`, `responsive_web_graphql_timeline_navigation_enabled`, `premium_content_api_read_enabled`, `communities_web_enable_tweet_community_results_fetch`, `c9s_tweet_anatomy_moderator_badge_enabled`, `responsive_web_grok_analyze_button_fetch_trends_enabled`, `responsive_web_grok_analyze_post_followups_enabled`, `rweb_cashtags_composer_attachment_enabled`, `responsive_web_jetfuel_frame`, `rweb_sports_post_context_enabled`, `responsive_web_grok_share_attachment_enabled`, `responsive_web_grok_annotations_enabled`, `articles_preview_enabled`, `responsive_web_edit_tweet_api_enabled`, `rweb_conversational_replies_downvote_enabled`, `graphql_is_translatable_rweb_tweet_is_translatable_enabled`, `view_counts_everywhere_api_enabled`, `longform_notetweets_consumption_enabled`, `responsive_web_twitter_article_tweet_consumption_enabled`, `content_disclosure_indicator_enabled`, `content_disclosure_ai_generated_indicator_enabled`, `responsive_web_grok_show_grok_translated_post`, `responsive_web_grok_analysis_button_from_backend`, `post_ctas_fetch_enabled`, `freedom_of_speech_not_reach_fetch_enabled`, `standardized_nudges_misinfo`, `tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled`, `longform_notetweets_rich_text_read_enabled`, `longform_notetweets_inline_media_enabled`, `responsive_web_nested_quote_preview_enabled`, `responsive_web_grok_image_annotation_enabled`, `responsive_web_grok_imagine_annotation_enabled`, `responsive_web_grok_community_note_auto_translation_is_enabled`, `responsive_web_enhance_cards_enabled` |
| `UserMedia` | 40 | `rweb_video_screen_enabled`, `rweb_cashtags_enabled`, `profile_label_improvements_pcf_label_in_post_enabled`, `responsive_web_profile_redirect_enabled`, `rweb_tipjar_consumption_enabled`, `verified_phone_label_enabled`, `creator_subscriptions_tweet_preview_api_enabled`, `responsive_web_graphql_timeline_navigation_enabled`, `premium_content_api_read_enabled`, `communities_web_enable_tweet_community_results_fetch`, `c9s_tweet_anatomy_moderator_badge_enabled`, `responsive_web_grok_analyze_button_fetch_trends_enabled`, `responsive_web_grok_analyze_post_followups_enabled`, `rweb_cashtags_composer_attachment_enabled`, `responsive_web_jetfuel_frame`, `rweb_sports_post_context_enabled`, `responsive_web_grok_share_attachment_enabled`, `responsive_web_grok_annotations_enabled`, `articles_preview_enabled`, `responsive_web_edit_tweet_api_enabled`, `rweb_conversational_replies_downvote_enabled`, `graphql_is_translatable_rweb_tweet_is_translatable_enabled`, `view_counts_everywhere_api_enabled`, `longform_notetweets_consumption_enabled`, `responsive_web_twitter_article_tweet_consumption_enabled`, `content_disclosure_indicator_enabled`, `content_disclosure_ai_generated_indicator_enabled`, `responsive_web_grok_show_grok_translated_post`, `responsive_web_grok_analysis_button_from_backend`, `post_ctas_fetch_enabled`, `freedom_of_speech_not_reach_fetch_enabled`, `standardized_nudges_misinfo`, `tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled`, `longform_notetweets_rich_text_read_enabled`, `longform_notetweets_inline_media_enabled`, `responsive_web_nested_quote_preview_enabled`, `responsive_web_grok_image_annotation_enabled`, `responsive_web_grok_imagine_annotation_enabled`, `responsive_web_grok_community_note_auto_translation_is_enabled`, `responsive_web_enhance_cards_enabled` |
| `UserOriginalsTimeline` | 40 | `rweb_video_screen_enabled`, `rweb_cashtags_enabled`, `profile_label_improvements_pcf_label_in_post_enabled`, `responsive_web_profile_redirect_enabled`, `rweb_tipjar_consumption_enabled`, `verified_phone_label_enabled`, `creator_subscriptions_tweet_preview_api_enabled`, `responsive_web_graphql_timeline_navigation_enabled`, `premium_content_api_read_enabled`, `communities_web_enable_tweet_community_results_fetch`, `c9s_tweet_anatomy_moderator_badge_enabled`, `responsive_web_grok_analyze_button_fetch_trends_enabled`, `responsive_web_grok_analyze_post_followups_enabled`, `rweb_cashtags_composer_attachment_enabled`, `responsive_web_jetfuel_frame`, `rweb_sports_post_context_enabled`, `responsive_web_grok_share_attachment_enabled`, `responsive_web_grok_annotations_enabled`, `articles_preview_enabled`, `responsive_web_edit_tweet_api_enabled`, `rweb_conversational_replies_downvote_enabled`, `graphql_is_translatable_rweb_tweet_is_translatable_enabled`, `view_counts_everywhere_api_enabled`, `longform_notetweets_consumption_enabled`, `responsive_web_twitter_article_tweet_consumption_enabled`, `content_disclosure_indicator_enabled`, `content_disclosure_ai_generated_indicator_enabled`, `responsive_web_grok_show_grok_translated_post`, `responsive_web_grok_analysis_button_from_backend`, `post_ctas_fetch_enabled`, `freedom_of_speech_not_reach_fetch_enabled`, `standardized_nudges_misinfo`, `tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled`, `longform_notetweets_rich_text_read_enabled`, `longform_notetweets_inline_media_enabled`, `responsive_web_nested_quote_preview_enabled`, `responsive_web_grok_image_annotation_enabled`, `responsive_web_grok_imagine_annotation_enabled`, `responsive_web_grok_community_note_auto_translation_is_enabled`, `responsive_web_enhance_cards_enabled` |
| `UserPhotoTimeline` | 40 | `rweb_video_screen_enabled`, `rweb_cashtags_enabled`, `profile_label_improvements_pcf_label_in_post_enabled`, `responsive_web_profile_redirect_enabled`, `rweb_tipjar_consumption_enabled`, `verified_phone_label_enabled`, `creator_subscriptions_tweet_preview_api_enabled`, `responsive_web_graphql_timeline_navigation_enabled`, `premium_content_api_read_enabled`, `communities_web_enable_tweet_community_results_fetch`, `c9s_tweet_anatomy_moderator_badge_enabled`, `responsive_web_grok_analyze_button_fetch_trends_enabled`, `responsive_web_grok_analyze_post_followups_enabled`, `rweb_cashtags_composer_attachment_enabled`, `responsive_web_jetfuel_frame`, `rweb_sports_post_context_enabled`, `responsive_web_grok_share_attachment_enabled`, `responsive_web_grok_annotations_enabled`, `articles_preview_enabled`, `responsive_web_edit_tweet_api_enabled`, `rweb_conversational_replies_downvote_enabled`, `graphql_is_translatable_rweb_tweet_is_translatable_enabled`, `view_counts_everywhere_api_enabled`, `longform_notetweets_consumption_enabled`, `responsive_web_twitter_article_tweet_consumption_enabled`, `content_disclosure_indicator_enabled`, `content_disclosure_ai_generated_indicator_enabled`, `responsive_web_grok_show_grok_translated_post`, `responsive_web_grok_analysis_button_from_backend`, `post_ctas_fetch_enabled`, `freedom_of_speech_not_reach_fetch_enabled`, `standardized_nudges_misinfo`, `tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled`, `longform_notetweets_rich_text_read_enabled`, `longform_notetweets_inline_media_enabled`, `responsive_web_nested_quote_preview_enabled`, `responsive_web_grok_image_annotation_enabled`, `responsive_web_grok_imagine_annotation_enabled`, `responsive_web_grok_community_note_auto_translation_is_enabled`, `responsive_web_enhance_cards_enabled` |
| `UserPreferences` | 0 | — |
| `UserPromotableTweets` | 40 | `rweb_video_screen_enabled`, `rweb_cashtags_enabled`, `profile_label_improvements_pcf_label_in_post_enabled`, `responsive_web_profile_redirect_enabled`, `rweb_tipjar_consumption_enabled`, `verified_phone_label_enabled`, `creator_subscriptions_tweet_preview_api_enabled`, `responsive_web_graphql_timeline_navigation_enabled`, `premium_content_api_read_enabled`, `communities_web_enable_tweet_community_results_fetch`, `c9s_tweet_anatomy_moderator_badge_enabled`, `responsive_web_grok_analyze_button_fetch_trends_enabled`, `responsive_web_grok_analyze_post_followups_enabled`, `rweb_cashtags_composer_attachment_enabled`, `responsive_web_jetfuel_frame`, `rweb_sports_post_context_enabled`, `responsive_web_grok_share_attachment_enabled`, `responsive_web_grok_annotations_enabled`, `articles_preview_enabled`, `responsive_web_edit_tweet_api_enabled`, `rweb_conversational_replies_downvote_enabled`, `graphql_is_translatable_rweb_tweet_is_translatable_enabled`, `view_counts_everywhere_api_enabled`, `longform_notetweets_consumption_enabled`, `responsive_web_twitter_article_tweet_consumption_enabled`, `content_disclosure_indicator_enabled`, `content_disclosure_ai_generated_indicator_enabled`, `responsive_web_grok_show_grok_translated_post`, `responsive_web_grok_analysis_button_from_backend`, `post_ctas_fetch_enabled`, `freedom_of_speech_not_reach_fetch_enabled`, `standardized_nudges_misinfo`, `tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled`, `longform_notetweets_rich_text_read_enabled`, `longform_notetweets_inline_media_enabled`, `responsive_web_nested_quote_preview_enabled`, `responsive_web_grok_image_annotation_enabled`, `responsive_web_grok_imagine_annotation_enabled`, `responsive_web_grok_community_note_auto_translation_is_enabled`, `responsive_web_enhance_cards_enabled` |
| `UserPromotedTweets` | 40 | `rweb_video_screen_enabled`, `rweb_cashtags_enabled`, `profile_label_improvements_pcf_label_in_post_enabled`, `responsive_web_profile_redirect_enabled`, `rweb_tipjar_consumption_enabled`, `verified_phone_label_enabled`, `creator_subscriptions_tweet_preview_api_enabled`, `responsive_web_graphql_timeline_navigation_enabled`, `premium_content_api_read_enabled`, `communities_web_enable_tweet_community_results_fetch`, `c9s_tweet_anatomy_moderator_badge_enabled`, `responsive_web_grok_analyze_button_fetch_trends_enabled`, `responsive_web_grok_analyze_post_followups_enabled`, `rweb_cashtags_composer_attachment_enabled`, `responsive_web_jetfuel_frame`, `rweb_sports_post_context_enabled`, `responsive_web_grok_share_attachment_enabled`, `responsive_web_grok_annotations_enabled`, `articles_preview_enabled`, `responsive_web_edit_tweet_api_enabled`, `rweb_conversational_replies_downvote_enabled`, `graphql_is_translatable_rweb_tweet_is_translatable_enabled`, `view_counts_everywhere_api_enabled`, `longform_notetweets_consumption_enabled`, `responsive_web_twitter_article_tweet_consumption_enabled`, `content_disclosure_indicator_enabled`, `content_disclosure_ai_generated_indicator_enabled`, `responsive_web_grok_show_grok_translated_post`, `responsive_web_grok_analysis_button_from_backend`, `post_ctas_fetch_enabled`, `freedom_of_speech_not_reach_fetch_enabled`, `standardized_nudges_misinfo`, `tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled`, `longform_notetweets_rich_text_read_enabled`, `longform_notetweets_inline_media_enabled`, `responsive_web_nested_quote_preview_enabled`, `responsive_web_grok_image_annotation_enabled`, `responsive_web_grok_imagine_annotation_enabled`, `responsive_web_grok_community_note_auto_translation_is_enabled`, `responsive_web_enhance_cards_enabled` |
| `UserRepliesTimeline` | 40 | `rweb_video_screen_enabled`, `rweb_cashtags_enabled`, `profile_label_improvements_pcf_label_in_post_enabled`, `responsive_web_profile_redirect_enabled`, `rweb_tipjar_consumption_enabled`, `verified_phone_label_enabled`, `creator_subscriptions_tweet_preview_api_enabled`, `responsive_web_graphql_timeline_navigation_enabled`, `premium_content_api_read_enabled`, `communities_web_enable_tweet_community_results_fetch`, `c9s_tweet_anatomy_moderator_badge_enabled`, `responsive_web_grok_analyze_button_fetch_trends_enabled`, `responsive_web_grok_analyze_post_followups_enabled`, `rweb_cashtags_composer_attachment_enabled`, `responsive_web_jetfuel_frame`, `rweb_sports_post_context_enabled`, `responsive_web_grok_share_attachment_enabled`, `responsive_web_grok_annotations_enabled`, `articles_preview_enabled`, `responsive_web_edit_tweet_api_enabled`, `rweb_conversational_replies_downvote_enabled`, `graphql_is_translatable_rweb_tweet_is_translatable_enabled`, `view_counts_everywhere_api_enabled`, `longform_notetweets_consumption_enabled`, `responsive_web_twitter_article_tweet_consumption_enabled`, `content_disclosure_indicator_enabled`, `content_disclosure_ai_generated_indicator_enabled`, `responsive_web_grok_show_grok_translated_post`, `responsive_web_grok_analysis_button_from_backend`, `post_ctas_fetch_enabled`, `freedom_of_speech_not_reach_fetch_enabled`, `standardized_nudges_misinfo`, `tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled`, `longform_notetweets_rich_text_read_enabled`, `longform_notetweets_inline_media_enabled`, `responsive_web_nested_quote_preview_enabled`, `responsive_web_grok_image_annotation_enabled`, `responsive_web_grok_imagine_annotation_enabled`, `responsive_web_grok_community_note_auto_translation_is_enabled`, `responsive_web_enhance_cards_enabled` |
| `UserRepostsTimeline` | 40 | `rweb_video_screen_enabled`, `rweb_cashtags_enabled`, `profile_label_improvements_pcf_label_in_post_enabled`, `responsive_web_profile_redirect_enabled`, `rweb_tipjar_consumption_enabled`, `verified_phone_label_enabled`, `creator_subscriptions_tweet_preview_api_enabled`, `responsive_web_graphql_timeline_navigation_enabled`, `premium_content_api_read_enabled`, `communities_web_enable_tweet_community_results_fetch`, `c9s_tweet_anatomy_moderator_badge_enabled`, `responsive_web_grok_analyze_button_fetch_trends_enabled`, `responsive_web_grok_analyze_post_followups_enabled`, `rweb_cashtags_composer_attachment_enabled`, `responsive_web_jetfuel_frame`, `rweb_sports_post_context_enabled`, `responsive_web_grok_share_attachment_enabled`, `responsive_web_grok_annotations_enabled`, `articles_preview_enabled`, `responsive_web_edit_tweet_api_enabled`, `rweb_conversational_replies_downvote_enabled`, `graphql_is_translatable_rweb_tweet_is_translatable_enabled`, `view_counts_everywhere_api_enabled`, `longform_notetweets_consumption_enabled`, `responsive_web_twitter_article_tweet_consumption_enabled`, `content_disclosure_indicator_enabled`, `content_disclosure_ai_generated_indicator_enabled`, `responsive_web_grok_show_grok_translated_post`, `responsive_web_grok_analysis_button_from_backend`, `post_ctas_fetch_enabled`, `freedom_of_speech_not_reach_fetch_enabled`, `standardized_nudges_misinfo`, `tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled`, `longform_notetweets_rich_text_read_enabled`, `longform_notetweets_inline_media_enabled`, `responsive_web_nested_quote_preview_enabled`, `responsive_web_grok_image_annotation_enabled`, `responsive_web_grok_imagine_annotation_enabled`, `responsive_web_grok_community_note_auto_translation_is_enabled`, `responsive_web_enhance_cards_enabled` |
| `UsersByRestIds` | 5 | `profile_label_improvements_pcf_label_in_post_enabled`, `responsive_web_profile_redirect_enabled`, `rweb_tipjar_consumption_enabled`, `verified_phone_label_enabled`, `responsive_web_graphql_timeline_navigation_enabled` |
| `UsersByScreenNames` | 12 | `hidden_profile_subscriptions_enabled`, `profile_label_improvements_pcf_label_in_post_enabled`, `responsive_web_profile_redirect_enabled`, `rweb_tipjar_consumption_enabled`, `verified_phone_label_enabled`, `subscriptions_verification_info_is_identity_verified_enabled`, `subscriptions_verification_info_verified_since_enabled`, `highlights_tweets_tab_ui_enabled`, `responsive_web_twitter_article_notes_tab_enabled`, `subscriptions_feature_can_gift_premium`, `creator_subscriptions_tweet_preview_api_enabled`, `responsive_web_graphql_timeline_navigation_enabled` |
| `UserSessionsList` | 0 | — |
| `UserSuperFollowTweets` | 40 | `rweb_video_screen_enabled`, `rweb_cashtags_enabled`, `profile_label_improvements_pcf_label_in_post_enabled`, `responsive_web_profile_redirect_enabled`, `rweb_tipjar_consumption_enabled`, `verified_phone_label_enabled`, `creator_subscriptions_tweet_preview_api_enabled`, `responsive_web_graphql_timeline_navigation_enabled`, `premium_content_api_read_enabled`, `communities_web_enable_tweet_community_results_fetch`, `c9s_tweet_anatomy_moderator_badge_enabled`, `responsive_web_grok_analyze_button_fetch_trends_enabled`, `responsive_web_grok_analyze_post_followups_enabled`, `rweb_cashtags_composer_attachment_enabled`, `responsive_web_jetfuel_frame`, `rweb_sports_post_context_enabled`, `responsive_web_grok_share_attachment_enabled`, `responsive_web_grok_annotations_enabled`, `articles_preview_enabled`, `responsive_web_edit_tweet_api_enabled`, `rweb_conversational_replies_downvote_enabled`, `graphql_is_translatable_rweb_tweet_is_translatable_enabled`, `view_counts_everywhere_api_enabled`, `longform_notetweets_consumption_enabled`, `responsive_web_twitter_article_tweet_consumption_enabled`, `content_disclosure_indicator_enabled`, `content_disclosure_ai_generated_indicator_enabled`, `responsive_web_grok_show_grok_translated_post`, `responsive_web_grok_analysis_button_from_backend`, `post_ctas_fetch_enabled`, `freedom_of_speech_not_reach_fetch_enabled`, `standardized_nudges_misinfo`, `tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled`, `longform_notetweets_rich_text_read_enabled`, `longform_notetweets_inline_media_enabled`, `responsive_web_nested_quote_preview_enabled`, `responsive_web_grok_image_annotation_enabled`, `responsive_web_grok_imagine_annotation_enabled`, `responsive_web_grok_community_note_auto_translation_is_enabled`, `responsive_web_enhance_cards_enabled` |
| `UsersVerifiedAvatars` | 1 | `responsive_web_graphql_timeline_navigation_enabled` |
| `UserTweets` | 40 | `rweb_video_screen_enabled`, `rweb_cashtags_enabled`, `profile_label_improvements_pcf_label_in_post_enabled`, `responsive_web_profile_redirect_enabled`, `rweb_tipjar_consumption_enabled`, `verified_phone_label_enabled`, `creator_subscriptions_tweet_preview_api_enabled`, `responsive_web_graphql_timeline_navigation_enabled`, `premium_content_api_read_enabled`, `communities_web_enable_tweet_community_results_fetch`, `c9s_tweet_anatomy_moderator_badge_enabled`, `responsive_web_grok_analyze_button_fetch_trends_enabled`, `responsive_web_grok_analyze_post_followups_enabled`, `rweb_cashtags_composer_attachment_enabled`, `responsive_web_jetfuel_frame`, `rweb_sports_post_context_enabled`, `responsive_web_grok_share_attachment_enabled`, `responsive_web_grok_annotations_enabled`, `articles_preview_enabled`, `responsive_web_edit_tweet_api_enabled`, `rweb_conversational_replies_downvote_enabled`, `graphql_is_translatable_rweb_tweet_is_translatable_enabled`, `view_counts_everywhere_api_enabled`, `longform_notetweets_consumption_enabled`, `responsive_web_twitter_article_tweet_consumption_enabled`, `content_disclosure_indicator_enabled`, `content_disclosure_ai_generated_indicator_enabled`, `responsive_web_grok_show_grok_translated_post`, `responsive_web_grok_analysis_button_from_backend`, `post_ctas_fetch_enabled`, `freedom_of_speech_not_reach_fetch_enabled`, `standardized_nudges_misinfo`, `tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled`, `longform_notetweets_rich_text_read_enabled`, `longform_notetweets_inline_media_enabled`, `responsive_web_nested_quote_preview_enabled`, `responsive_web_grok_image_annotation_enabled`, `responsive_web_grok_imagine_annotation_enabled`, `responsive_web_grok_community_note_auto_translation_is_enabled`, `responsive_web_enhance_cards_enabled` |
| `UserTweetsAndReplies` | 40 | `rweb_video_screen_enabled`, `rweb_cashtags_enabled`, `profile_label_improvements_pcf_label_in_post_enabled`, `responsive_web_profile_redirect_enabled`, `rweb_tipjar_consumption_enabled`, `verified_phone_label_enabled`, `creator_subscriptions_tweet_preview_api_enabled`, `responsive_web_graphql_timeline_navigation_enabled`, `premium_content_api_read_enabled`, `communities_web_enable_tweet_community_results_fetch`, `c9s_tweet_anatomy_moderator_badge_enabled`, `responsive_web_grok_analyze_button_fetch_trends_enabled`, `responsive_web_grok_analyze_post_followups_enabled`, `rweb_cashtags_composer_attachment_enabled`, `responsive_web_jetfuel_frame`, `rweb_sports_post_context_enabled`, `responsive_web_grok_share_attachment_enabled`, `responsive_web_grok_annotations_enabled`, `articles_preview_enabled`, `responsive_web_edit_tweet_api_enabled`, `rweb_conversational_replies_downvote_enabled`, `graphql_is_translatable_rweb_tweet_is_translatable_enabled`, `view_counts_everywhere_api_enabled`, `longform_notetweets_consumption_enabled`, `responsive_web_twitter_article_tweet_consumption_enabled`, `content_disclosure_indicator_enabled`, `content_disclosure_ai_generated_indicator_enabled`, `responsive_web_grok_show_grok_translated_post`, `responsive_web_grok_analysis_button_from_backend`, `post_ctas_fetch_enabled`, `freedom_of_speech_not_reach_fetch_enabled`, `standardized_nudges_misinfo`, `tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled`, `longform_notetweets_rich_text_read_enabled`, `longform_notetweets_inline_media_enabled`, `responsive_web_nested_quote_preview_enabled`, `responsive_web_grok_image_annotation_enabled`, `responsive_web_grok_imagine_annotation_enabled`, `responsive_web_grok_community_note_auto_translation_is_enabled`, `responsive_web_enhance_cards_enabled` |
| `UserVideoTimeline` | 40 | `rweb_video_screen_enabled`, `rweb_cashtags_enabled`, `profile_label_improvements_pcf_label_in_post_enabled`, `responsive_web_profile_redirect_enabled`, `rweb_tipjar_consumption_enabled`, `verified_phone_label_enabled`, `creator_subscriptions_tweet_preview_api_enabled`, `responsive_web_graphql_timeline_navigation_enabled`, `premium_content_api_read_enabled`, `communities_web_enable_tweet_community_results_fetch`, `c9s_tweet_anatomy_moderator_badge_enabled`, `responsive_web_grok_analyze_button_fetch_trends_enabled`, `responsive_web_grok_analyze_post_followups_enabled`, `rweb_cashtags_composer_attachment_enabled`, `responsive_web_jetfuel_frame`, `rweb_sports_post_context_enabled`, `responsive_web_grok_share_attachment_enabled`, `responsive_web_grok_annotations_enabled`, `articles_preview_enabled`, `responsive_web_edit_tweet_api_enabled`, `rweb_conversational_replies_downvote_enabled`, `graphql_is_translatable_rweb_tweet_is_translatable_enabled`, `view_counts_everywhere_api_enabled`, `longform_notetweets_consumption_enabled`, `responsive_web_twitter_article_tweet_consumption_enabled`, `content_disclosure_indicator_enabled`, `content_disclosure_ai_generated_indicator_enabled`, `responsive_web_grok_show_grok_translated_post`, `responsive_web_grok_analysis_button_from_backend`, `post_ctas_fetch_enabled`, `freedom_of_speech_not_reach_fetch_enabled`, `standardized_nudges_misinfo`, `tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled`, `longform_notetweets_rich_text_read_enabled`, `longform_notetweets_inline_media_enabled`, `responsive_web_nested_quote_preview_enabled`, `responsive_web_grok_image_annotation_enabled`, `responsive_web_grok_imagine_annotation_enabled`, `responsive_web_grok_community_note_auto_translation_is_enabled`, `responsive_web_enhance_cards_enabled` |
| `Viewer` | 7 | `subscriptions_upsells_api_enabled`, `profile_label_improvements_pcf_label_in_post_enabled`, `responsive_web_profile_redirect_enabled`, `rweb_tipjar_consumption_enabled`, `verified_phone_label_enabled`, `creator_subscriptions_tweet_preview_api_enabled`, `responsive_web_graphql_timeline_navigation_enabled` |
| `ViewerBadgeCounts` | 0 | — |
| `WriteDataSaverPreferences` | 0 | — |

## Valeurs des indicateurs découverts

Valeurs booléennes extraites du HTML authentifié au moment du relevé. Elles décrivent la configuration observée pour cette session X, pas une garantie de disponibilité des fonctions correspondantes.

| Indicateur | Valeur |
| --- | --- |
| `2fa_temporary_password_enabled` | `false` |
| `active_ad_campaigns_query_enabled` | `false` |
| `android_ui_nested_quote_tweet_preview_enabled` | `false` |
| `Arkose_rweb_hosted_page` | `true` |
| `Arkose_use_invisible_challenge_key` | `false` |
| `articles_preview_enabled` | `true` |
| `articles_rest_api_enabled` | `true` |
| `av_chat_encryption_enabled` | `false` |
| `av_chat_group_e2ee_creator_enabled` | `false` |
| `av_chat_group_e2ee_joiner_enabled` | `false` |
| `av_chat_xchat_emoji_reactions_enabled` | `false` |
| `birdwatch_xchat_pivot_rendering_enabled` | `true` |
| `blue_business_admin_sidebar_module_enabled` | `true` |
| `blue_business_ads_metrics` | `true` |
| `blue_business_affiliates_list_order_setting_enabled` | `false` |
| `blue_business_analytics` | `true` |
| `blue_business_analytics_affiliate_filtering_enabled` | `true` |
| `blue_business_direct_invites_enabled` | `true` |
| `blue_business_display_annual_price_monthly` | `true` |
| `blue_business_multi_affiliates_ui_enabled` | `true` |
| `blue_business_simplify_signup_ui` | `false` |
| `blue_business_tier_switching_enabled` | `true` |
| `blue_business_username_change_prompt_enabled` | `true` |
| `blue_business_verified_admin_enabled` | `true` |
| `blue_business_vo_nav_for_legacy_verified` | `false` |
| `blue_longer_video_enabled` | `false` |
| `branded_features_is_branded_likes_on_tweet_content_enabled` | `true` |
| `branded_features_search_overlay_animations_enabled` | `false` |
| `branded_like_preview_enabled` | `false` |
| `broadcast_live_chat_enabled` | `true` |
| `c9s_auto_collapse_community_detail_header_enabled` | `true` |
| `c9s_community_answer_box_enabled` | `true` |
| `c9s_community_answer_box_join_page_enabled` | `true` |
| `c9s_community_composer_hashtag_suggestions_enabled` | `true` |
| `c9s_community_hashtags_carousel_enabled` | `true` |
| `c9s_community_hashtags_enabled` | `true` |
| `c9s_community_list_setting_enabled` | `true` |
| `c9s_community_question_box_enabled` | `true` |
| `c9s_community_searchtags_enabled` | `true` |
| `c9s_community_tweet_search_enabled` | `true` |
| `c9s_enabled` | `true` |
| `c9s_list_members_action_api_enabled` | `false` |
| `c9s_nav_list_activity_details_enabled` | `false` |
| `c9s_question_editing_box_enabled` | `true` |
| `c9s_spotlight_creation_enabled` | `true` |
| `c9s_timelines_media_tab_enabled` | `true` |
| `c9s_tweet_anatomy_moderator_badge_enabled` | `true` |
| `co_timeline_topic_filter_enabled` | `false` |
| `communities_adult_content_setting_display` | `true` |
| `communities_adult_content_setting_enabled` | `true` |
| `communities_analytics_enabled` | `true` |
| `communities_auto_report_setting_enabled` | `true` |
| `communities_enable_explore_tab` | `true` |
| `communities_enable_explore_topic_carousel` | `true` |
| `communities_enable_top_posts_search` | `true` |
| `communities_global_communities_latest_post_search_enabled` | `true` |
| `communities_global_communities_post_search_enabled` | `true` |
| `communities_home_top_timeline_enabled` | `true` |
| `communities_moderation_log_enabled` | `true` |
| `communities_non_member_reply_enabled` | `true` |
| `communities_show_broadcast_option_in_composer` | `true` |
| `communities_spam_settings_enabled` | `true` |
| `communities_topic_carousel_enabled` | `true` |
| `communities_topic_display` | `true` |
| `communities_topics_enabled` | `true` |
| `communities_web_enable_tweet_community_results_fetch` | `true` |
| `complat_management_enabled` | `false` |
| `complat_responsive_web_castle_minting_enabled` | `false` |
| `complat_x_migration_enabled` | `false` |
| `composer_add_geo_network_conversation_controls` | `true` |
| `composer_composition_signals_enabled` | `true` |
| `content_disclosure_ai_generated_c2pa_detection_enabled` | `true` |
| `content_disclosure_ai_generated_creation_enabled` | `true` |
| `content_disclosure_ai_generated_indicator_enabled` | `true` |
| `content_disclosure_creation_enabled` | `true` |
| `content_disclosure_indicator_enabled` | `true` |
| `creator_monetization_profile_subscription_tweets_tab_enabled` | `true` |
| `creator_studio_nav_enabled` | `true` |
| `creator_subscriptions_connect_tab_enabled` | `true` |
| `creator_subscriptions_email_share_enabled` | `true` |
| `creator_subscriptions_revamp_enabled` | `true` |
| `creator_subscriptions_subscribe_action_tweet_menu_enabled` | `true` |
| `creator_subscriptions_subscribe_button_tweet_detail_enabled` | `true` |
| `creator_subscriptions_subscriber_count_enabled` | `false` |
| `creator_subscriptions_subscription_count_enabled` | `true` |
| `creator_subscriptions_tweet_preview_api_enabled` | `true` |
| `dash_region_specific_de_and_tr_media_transparency_items_enabled` | `false` |
| `dash_region_specific_de_media_transparency_items_enabled` | `false` |
| `disallowed_reply_controls_callout_enabled` | `false` |
| `disallowed_reply_controls_enabled` | `false` |
| `dm_block_enabled` | `true` |
| `dm_bulk_delete_enabled` | `false` |
| `dm_conversation_labels_pinned_education_enabled` | `true` |
| `dm_conversations_nsfw_media_filter_enabled` | `false` |
| `dm_edit_dms_overflow_menu_enabled` | `false` |
| `dm_education_flags_prompt` | `false` |
| `dm_secret_conversations_enabled` | `false` |
| `dm_settings_info_page_allow_subscriber_messages_setting_enabled` | `true` |
| `dm_settings_info_page_device_list_enabled` | `false` |
| `dm_video_downloads_enabled` | `false` |
| `dm_voice_rendering_enabled` | `true` |
| `dont_mention_me_enabled` | `true` |
| `dont_mention_me_mentions_tab_education_enabled` | `true` |
| `dont_mention_me_view_api_enabled` | `true` |
| `dsa_encrypted_dms_report_flow_enabled` | `false` |
| `dsa_profile_report_flow_enabled` | `false` |
| `dsa_report_flow_enabled` | `false` |
| `ecd_dispute_form_link_enabled` | `true` |
| `enable_label_appealing_misinfo_enabled` | `false` |
| `enable_label_appealing_sensitive_content_enabled` | `false` |
| `explore_graphql_enabled` | `true` |
| `explore_relaunch_enable_auto_play` | `false` |
| `explore_relaunch_enable_immersive_web` | `false` |
| `explore_relaunch_enable_immersive_web_navigation_button` | `false` |
| `follow_nudge_conversation_enabled` | `false` |
| `freedom_of_speech_not_reach_author_label_enabled` | `true` |
| `freedom_of_speech_not_reach_fetch_enabled` | `true` |
| `freedom_of_speech_not_reach_pivot_enabled` | `true` |
| `graduated_access_botmaker_decider_enabled` | `true` |
| `graduated_access_invisible_treatment_enabled` | `true` |
| `graduated_access_user_prompt_enabled` | `true` |
| `graphql_is_translatable_rweb_tweet_is_translatable_enabled` | `true` |
| `grok_agent_disclosure_enabled` | `false` |
| `grok_settings_age_restriction_enabled` | `true` |
| `grok_translations_community_note_auto_translation_is_enabled` | `false` |
| `grok_translations_community_note_translation_is_enabled` | `false` |
| `grok_translations_post_auto_translation_is_enabled` | `false` |
| `grok_xweb_connectors_enabled` | `false` |
| `grok_xweb_skills_enabled` | `false` |
| `gryphon_faster_cell_entrance` | `true` |
| `gryphon_fps_tracking_enabled` | `true` |
| `gryphon_hide_quick_promote` | `false` |
| `gryphon_live_timelines_enabled` | `true` |
| `gryphon_motion` | `false` |
| `gryphon_redux_perf_optimization_enabled` | `true` |
| `gryphon_redux_perf_optimization_v2_enabled` | `true` |
| `gryphon_search_based_deck_enabled` | `false` |
| `gryphon_survey_enabled` | `false` |
| `gryphon_underground_enabled` | `false` |
| `gryphon_upgrade_premium_plus_banner_enabled` | `true` |
| `gryphon_video_docking_enabled` | `true` |
| `hashfetti_all_hashflags` | `false` |
| `hashfetti_also_match_query` | `false` |
| `hashfetti_enabled` | `true` |
| `hidden_profile_subscriptions_enabled` | `true` |
| `highlights_tweets_action_enabled` | `true` |
| `highlights_tweets_action_menu_upsell_enabled` | `true` |
| `highlights_tweets_tab_ui_enabled` | `true` |
| `highlights_tweets_tab_upsell_enabled` | `true` |
| `highlights_tweets_upsell_on_pin_action_enabled` | `false` |
| `home_timeline_like_reactivity_enabled` | `true` |
| `home_timeline_spheres_detail_page_muting_enabled` | `true` |
| `home_timeline_spheres_ranking_mode_control_enabled` | `false` |
| `hometimeline_pinned_tabs_topics_enabled` | `true` |
| `identity_verification_consent_opt_in_by_default_enabled` | `true` |
| `identity_verification_debadging_notification_enabled` | `true` |
| `identity_verification_hide_verified_label_settings_enabled` | `true` |
| `identity_verification_intake_enabled` | `false` |
| `identity_verification_intake_for_blue_subscribers_enabled` | `false` |
| `identity_verification_notable_demo_survey` | `false` |
| `identity_verification_passkey_settings_enabled` | `true` |
| `identity_verification_settings_enabled` | `true` |
| `identity_verification_vendor_idv_migration_enabled` | `false` |
| `immersive_video_status_linkable_timestamps` | `false` |
| `immersive_viewer_enable_profile_viewer` | `false` |
| `insights_ai_trends_enabled` | `false` |
| `insights_chart_filter_enabled` | `true` |
| `insights_paginated_metrics_backend_enabled` | `false` |
| `insights_preview_splash_metrics_enabled` | `false` |
| `insights_previews_enabled` | `false` |
| `ios_premium_paywall_preloaded_webview_pagesheet_modal` | `true` |
| `jetfuel_forgot_password_migration_enabled` | `true` |
| `krs_registration_enabled` | `false` |
| `live_event_docking_enabled` | `true` |
| `live_event_interstitial_seen_cache_enabled` | `true` |
| `live_event_multi_video_auto_advance_dock_enabled` | `true` |
| `live_event_multi_video_auto_advance_enabled` | `true` |
| `live_event_multi_video_auto_advance_fullscreen_enabled` | `false` |
| `live_event_multi_video_enabled` | `true` |
| `live_event_timeline_server_controlled_refresh_rate_enabled` | `true` |
| `livepipeline_client_enabled` | `true` |
| `livepipeline_tweetengagement_enabled` | `true` |
| `longform_notetweets_composer_upsell_enabled` | `true` |
| `longform_notetweets_composition_without_claims_enabled` | `false` |
| `longform_notetweets_consumption_enabled` | `true` |
| `longform_notetweets_inline_media_enabled` | `false` |
| `longform_notetweets_mobile_richtextinput` | `false` |
| `longform_notetweets_rich_text_read_enabled` | `true` |
| `longform_notetweets_rich_text_timeline_enabled` | `false` |
| `longform_notetweets_scheduling_non_reply_enabled` | `true` |
| `longform_notetweets_tweet_storm_enabled` | `true` |
| `longform_reader_mode_view_in_reader_mode_entry_button_enabled` | `false` |
| `machine_translation_holdback_logged_in` | `false` |
| `machine_translation_holdback_logged_out` | `false` |
| `march_madness_brackets_enabled` | `true` |
| `march_madness_brackets_enabled_loggedin_sidebar_popup` | `false` |
| `media_async_upload_longer_video_resolution_selector` | `false` |
| `media_edge_to_edge_content_enabled` | `false` |
| `netzdg_in_spaces_enabled` | `false` |
| `new_timeline_experiment_enabled` | `false` |
| `oauth_consent_account_switch_enabled` | `true` |
| `oauth_trusted_developer_badge_enabled` | `true` |
| `ocf_2fa_enrollment_bouncer_enabled` | `true` |
| `ocf_2fa_enrollment_enabled` | `true` |
| `ocf_2fa_unenrollment_enabled` | `true` |
| `onboarding_project_uls_enabled` | `true` |
| `papago_tweet_translation_from_korean_entity_protected` | `false` |
| `papago_tweet_translation_from_korean_entity_unprotected` | `false` |
| `papago_tweet_translation_to_korean` | `false` |
| `payments_1password_history_fix_enabled` | `true` |
| `payments_agent_connections_enabled` | `true` |
| `payments_agent_connections_prefill_enabled` | `false` |
| `payments_auth_token_enabled` | `true` |
| `payments_cash_deposits_enabled` | `true` |
| `payments_chat_support_enabled` | `false` |
| `payments_chat_support_for_limits_enabled` | `true` |
| `payments_cheques_deposits_enabled` | `true` |
| `payments_csv_export_enabled` | `false` |
| `payments_forward_with_enabled` | `true` |
| `payments_funds_insurance_redesign_enabled` | `true` |
| `payments_funds_insurance_sweep_enabled` | `true` |
| `payments_half_cover_notices_enabled` | `true` |
| `payments_interest_details_with_boost_enabled` | `true` |
| `payments_international_wires_enabled` | `false` |
| `payments_lexical_search_enabled` | `false` |
| `payments_onboarding_issued_cards_polling_enabled` | `true` |
| `payments_passkey_onboarding_enabled` | `true` |
| `payments_scheduled_payments_enabled` | `true` |
| `payments_secondary_accounts_enabled` | `true` |
| `payments_shared_accounts_enabled` | `true` |
| `payments_support_grpc_enabled` | `true` |
| `payments_support_xchat_sessions_enabled` | `true` |
| `payments_timeline_action` | `true` |
| `payments_tracing_reports_enabled` | `true` |
| `payments_transaction_receipt_enabled` | `true` |
| `payments_transaction_search_enabled` | `true` |
| `payments_transfer_link_enabled` | `false` |
| `payments_web_external_app_enabled` | `false` |
| `people_search_interests_filter_enabled` | `false` |
| `post_ctas_fetch_enabled` | `false` |
| `post_ctas_render_enabled` | `false` |
| `premium_business_offers_banner_portal_basic_tier` | `false` |
| `premium_business_offers_banner_sidebar_basic_tier` | `false` |
| `premium_business_offers_nav_indicator_enabled` | `false` |
| `premium_business_offers_navbar_discount_label_enabled` | `false` |
| `premium_business_offers_navbar_premium_signup_hidden` | `false` |
| `premium_business_offers_signup_navbar_tab_enabled` | `false` |
| `premium_content_api_read_enabled` | `false` |
| `premium_paywall_on_app_load_enabled` | `false` |
| `premium_paywall_on_app_load_journey_enabled` | `true` |
| `premium_webview_paywall_custom_timelines_feature_enabled` | `true` |
| `premium_webview_paywall_force_premium_tier_enabled` | `false` |
| `premium_webview_paywall_intro_offer_title_new_copy_enabled` | `true` |
| `premium_webview_paywall_tier_switch_all_plans_button_hidden` | `true` |
| `premium_webview_paywall_tier_switch_upgrade_disclaimer_enabled` | `true` |
| `professional_launchpad_m1_enabled` | `true` |
| `professional_launchpad_mobile_promotable_timeline` | `false` |
| `professional_launchpad_upload_address_book` | `true` |
| `profile_label_improvements_pcf_edit_profile_enabled` | `true` |
| `profile_label_improvements_pcf_label_in_post_enabled` | `true` |
| `profile_label_improvements_pcf_settings_enabled` | `true` |
| `recruiting_admin_currencies_enabled` | `false` |
| `recruiting_global_jobs_search_enabled` | `false` |
| `recruiting_job_page_consumption_enabled` | `false` |
| `recruiting_job_recommendations_enabled` | `false` |
| `recruiting_job_search_ai_companies_filter_enabled` | `false` |
| `recruiting_jobs_list_consumption_enabled` | `false` |
| `recruiting_jobs_list_search_enabled` | `false` |
| `recruiting_jobs_list_share_enabled` | `false` |
| `recruiting_pin_job_enabled` | `false` |
| `recruiting_premium_jobs_enabled` | `false` |
| `recruiting_promoted_jobs_enabled` | `false` |
| `recruiting_search_filters_enabled` | `false` |
| `recruiting_verified_orgs_admin_enabled` | `false` |
| `recruiting_verified_orgs_ats_integration_enabled` | `false` |
| `recruiting_verified_orgs_enroll_allowed` | `false` |
| `report_center_mvp_r1_enabled` | `true` |
| `report_center_mvp_r2_enabled` | `false` |
| `responsive_web_ad_formats_enable_dismiss_in_home_urt` | `true` |
| `responsive_web_ad_formats_hide_vanity_for_business_account` | `false` |
| `responsive_web_ad_formats_media_overlay_enabled` | `true` |
| `responsive_web_ad_formats_website_cta_enabled` | `true` |
| `responsive_web_ad_revenue_sharing_bounce_all_legacy_to_creator_studio_enabled` | `true` |
| `responsive_web_ad_revenue_sharing_dashboard_redirect_enabled` | `false` |
| `responsive_web_ad_revenue_sharing_enabled` | `true` |
| `responsive_web_ad_revenue_sharing_onboarding_redirect_enabled` | `false` |
| `responsive_web_ad_revenue_sharing_setup_enabled` | `false` |
| `responsive_web_ad_revenue_sharing_subscriptions_dashboard_redirect_enabled` | `false` |
| `responsive_web_ad_revenue_sharing_total_earnings_enabled` | `false` |
| `responsive_web_ad_revenue_sharing_url_update_enabled` | `true` |
| `responsive_web_alt_text_nudges_enabled` | `true` |
| `responsive_web_alt_text_nudges_settings_enabled` | `true` |
| `responsive_web_alt_text_translations_enabled` | `true` |
| `responsive_web_android_quick_promote_stripe_elements_enabled` | `false` |
| `responsive_web_api_transition_enabled` | `true` |
| `responsive_web_audio_space_ring_home_timeline` | `false` |
| `responsive_web_author_labels_avatar_label_enabled` | `false` |
| `responsive_web_author_labels_focal_label_enabled` | `false` |
| `responsive_web_author_labels_handle_label_enabled` | `false` |
| `responsive_web_birdwatch_admitted_user_setting_enabled` | `false` |
| `responsive_web_birdwatch_consumption_enabled` | `true` |
| `responsive_web_birdwatch_country_allowed` | `true` |
| `responsive_web_birdwatch_crowd_nmr_pivots_enabled` | `true` |
| `responsive_web_birdwatch_enforce_author_user_quotas` | `true` |
| `responsive_web_birdwatch_fast_notes_badge_enabled` | `false` |
| `responsive_web_birdwatch_home_page_enabled` | `false` |
| `responsive_web_birdwatch_live_note_classification_enabled` | `false` |
| `responsive_web_birdwatch_live_note_enabled` | `true` |
| `responsive_web_birdwatch_match_page_enabled` | `true` |
| `responsive_web_birdwatch_media_history_enabled` | `false` |
| `responsive_web_birdwatch_media_notes_enabled` | `true` |
| `responsive_web_birdwatch_netzdg_enabled` | `false` |
| `responsive_web_birdwatch_note_internal_insights_enabled` | `false` |
| `responsive_web_birdwatch_note_limit_enabled` | `true` |
| `responsive_web_birdwatch_note_request_download_enabled` | `true` |
| `responsive_web_birdwatch_note_request_enabled` | `true` |
| `responsive_web_birdwatch_note_request_sources_enabled` | `true` |
| `responsive_web_birdwatch_note_request_suggestion_enabled` | `true` |
| `responsive_web_birdwatch_note_writing_enabled` | `false` |
| `responsive_web_birdwatch_notification_settings_enabled` | `true` |
| `responsive_web_birdwatch_pivots_enabled` | `true` |
| `responsive_web_birdwatch_public_suggestions_tab_enabled` | `true` |
| `responsive_web_birdwatch_rating_crowd_enabled` | `true` |
| `responsive_web_birdwatch_rating_participant_enabled` | `false` |
| `responsive_web_birdwatch_require_rating_before_writing_enabled` | `true` |
| `responsive_web_birdwatch_self_remove_enabled` | `true` |
| `responsive_web_birdwatch_show_note_request_sources_despite_top_writer_status` | `false` |
| `responsive_web_birdwatch_signup_prompt_enabled` | `true` |
| `responsive_web_birdwatch_site_enabled` | `true` |
| `responsive_web_birdwatch_suggestion_rating_impact_enabled` | `true` |
| `responsive_web_birdwatch_suggestions_report_enabled` | `true` |
| `responsive_web_birdwatch_top_contributor_enabled` | `true` |
| `responsive_web_birdwatch_translation_enabled` | `true` |
| `responsive_web_birdwatch_url_notes_enabled` | `false` |
| `responsive_web_card_image_poll_enabled` | `true` |
| `responsive_web_card_image_poll_shuffle_enabled` | `true` |
| `responsive_web_card_image_poll_sort_by_vote_count_enabled` | `true` |
| `responsive_web_card_preconnect_enabled` | `false` |
| `responsive_web_card_reminder_enabled` | `false` |
| `responsive_web_carousel_v2_media_detail_enabled` | `false` |
| `responsive_web_castle_client_event_enabled` | `false` |
| `responsive_web_castle_sdk_enabled` | `true` |
| `responsive_web_chat_enabled` | `true` |
| `responsive_web_commerce_shop_spotlight_enabled` | `false` |
| `responsive_web_communityboost_download_data_enabled` | `false` |
| `responsive_web_communityboost_form_enabled` | `false` |
| `responsive_web_communityboost_mixed_pivot_enabled` | `false` |
| `responsive_web_composer_autosave_enabled` | `false` |
| `responsive_web_composer_configurable_video_player_enabled` | `false` |
| `responsive_web_convert_card_video_to_gif_enabled` | `false` |
| `responsive_web_cookie_compliance_banner_enabled` | `false` |
| `responsive_web_cookie_compliance_banner_update_enabled` | `false` |
| `responsive_web_cookie_compliance_gingersnap_enabled` | `false` |
| `responsive_web_cookie_consent_signal_enabled` | `false` |
| `responsive_web_creator_preferences_previews_enabled_setting` | `true` |
| `responsive_web_dcm_2_enabled` | `true` |
| `responsive_web_device_follow_without_user_follow_enabled` | `false` |
| `responsive_web_disconnect_third_party_sso_enabled` | `true` |
| `responsive_web_dockable_autoplay_policy_enabled` | `true` |
| `responsive_web_edit_tweet_api_enabled` | `true` |
| `responsive_web_edit_tweet_composition_enabled` | `true` |
| `responsive_web_edit_tweet_enabled` | `false` |
| `responsive_web_edit_tweet_perspective_enabled` | `false` |
| `responsive_web_edit_tweet_upsell_enabled` | `true` |
| `responsive_web_element_size_impression_scribe_enabled` | `true` |
| `responsive_web_enhance_cards_enabled` | `false` |
| `responsive_web_extension_compatibility_hide` | `false` |
| `responsive_web_extension_compatibility_impression_guard` | `true` |
| `responsive_web_extension_compatibility_override_param` | `false` |
| `responsive_web_extension_compatibility_scribe` | `true` |
| `responsive_web_fetch_hashflags_on_boot` | `false` |
| `responsive_web_framerate_tracking_home_enabled` | `false` |
| `responsive_web_gpc_enabled` | `false` |
| `responsive_web_graphql_feedback` | `true` |
| `responsive_web_graphql_skip_user_profile_image_extensions_enabled` | `false` |
| `responsive_web_graphql_timeline_navigation_enabled` | `true` |
| `responsive_web_grok_05221996` | `false` |
| `responsive_web_grok_420_toggle_enabled` | `false` |
| `responsive_web_grok_allow_youtube_embeds` | `false` |
| `responsive_web_grok_analysis_button_from_backend` | `true` |
| `responsive_web_grok_analyze_button_fetch_trends_enabled` | `false` |
| `responsive_web_grok_analyze_focal_post_enabled` | `false` |
| `responsive_web_grok_analyze_post_followups_enabled` | `false` |
| `responsive_web_grok_annotations_enabled` | `true` |
| `responsive_web_grok_api_enable_grok_host` | `true` |
| `responsive_web_grok_article_cover_image_gen_enabled` | `false` |
| `responsive_web_grok_article_summary_enabled` | `true` |
| `responsive_web_grok_backend_prompts_enabled` | `true` |
| `responsive_web_grok_bio_auto_translation_in_followers_enabled` | `true` |
| `responsive_web_grok_bio_auto_translation_in_search_is_enabled` | `true` |
| `responsive_web_grok_bio_auto_translation_is_enabled` | `true` |
| `responsive_web_grok_bot_preset_enabled` | `false` |
| `responsive_web_grok_build_tab_relabel_enabled` | `false` |
| `responsive_web_grok_community_note_auto_translation_is_enabled` | `true` |
| `responsive_web_grok_community_note_translation_is_enabled` | `true` |
| `responsive_web_grok_debug_enabled` | `false` |
| `responsive_web_grok_dev_universal_search_id_enabled` | `false` |
| `responsive_web_grok_disable_new_conversation_url_reset` | `false` |
| `responsive_web_grok_download_cta_enabled` | `true` |
| `responsive_web_grok_download_favicons` | `true` |
| `responsive_web_grok_enable_android_image_donwload` | `false` |
| `responsive_web_grok_enable_deepersearch` | `true` |
| `responsive_web_grok_enable_grok_analyze_education` | `false` |
| `responsive_web_grok_enable_grok_tab_education` | `false` |
| `responsive_web_grok_enable_video_gen_on_image_preview` | `false` |
| `responsive_web_grok_fade_in_animation_v2_enabled` | `true` |
| `responsive_web_grok_feed` | `false` |
| `responsive_web_grok_file_upload_enabled` | `true` |
| `responsive_web_grok_fun_mode_disabled` | `true` |
| `responsive_web_grok_general_availability` | `false` |
| `responsive_web_grok_home_dark_enabled` | `true` |
| `responsive_web_grok_image_annotation_enabled` | `true` |
| `responsive_web_grok_image_edit` | `true` |
| `responsive_web_grok_image_lazyload_enabled` | `true` |
| `responsive_web_grok_imagine_annotation_enabled` | `true` |
| `responsive_web_grok_imagine_composer_enabled` | `false` |
| `responsive_web_grok_imagine_continue_seamless_enabled` | `true` |
| `responsive_web_grok_imagine_explore_enabled` | `false` |
| `responsive_web_grok_imagine_image_comparison_enabled` | `false` |
| `responsive_web_grok_imagine_in_composer_enabled` | `false` |
| `responsive_web_grok_imagine_native_share_enabled` | `false` |
| `responsive_web_grok_imagine_profile_edit_enabled` | `false` |
| `responsive_web_grok_img_composer` | `true` |
| `responsive_web_grok_img_composer_in_media_picker` | `false` |
| `responsive_web_grok_img_composer_redirect_to_grok_com_imagine` | `true` |
| `responsive_web_grok_inplace_auth_handoff_enabled` | `true` |
| `responsive_web_grok_latest_news_preset_enabled` | `true` |
| `responsive_web_grok_link_edit_image_to_grok_com_enabled` | `true` |
| `responsive_web_grok_location_enabled` | `true` |
| `responsive_web_grok_media_attribution_focal_post_force_show` | `false` |
| `responsive_web_grok_media_attribution_imagine_force_show` | `false` |
| `responsive_web_grok_media_attribution_route_to_imagine_composer` | `false` |
| `responsive_web_grok_media_block_edit_enabled` | `true` |
| `responsive_web_grok_model_selector_in_input` | `true` |
| `responsive_web_grok_personality` | `true` |
| `responsive_web_grok_places_card_enabled` | `false` |
| `responsive_web_grok_post_inline_translation_is_enabled` | `true` |
| `responsive_web_grok_post_understanding_button_on_all_posts` | `false` |
| `responsive_web_grok_profile_summary_enabled` | `true` |
| `responsive_web_grok_promo_modal_enabled` | `false` |
| `responsive_web_grok_prompt_edit_enabled` | `true` |
| `responsive_web_grok_redirect_enabled` | `true` |
| `responsive_web_grok_regen_configs` | `false` |
| `responsive_web_grok_route_disabled_search_think_to_paywall` | `true` |
| `responsive_web_grok_rtl_detection` | `true` |
| `responsive_web_grok_search_summary_enabled` | `false` |
| `responsive_web_grok_search_summary_images_enabled` | `true` |
| `responsive_web_grok_search_summary_sidebar` | `true` |
| `responsive_web_grok_share_attachment_enabled` | `true` |
| `responsive_web_grok_show_button_is_ad` | `false` |
| `responsive_web_grok_show_button_on_ads` | `false` |
| `responsive_web_grok_show_button_send_is_ads` | `false` |
| `responsive_web_grok_show_cards_at_top` | `true` |
| `responsive_web_grok_show_citations` | `true` |
| `responsive_web_grok_show_grok_performance_metrics` | `false` |
| `responsive_web_grok_show_grok_translated_post` | `true` |
| `responsive_web_grok_show_message_post_button` | `true` |
| `responsive_web_grok_sport_cards_enabled` | `true` |
| `responsive_web_grok_start_title_experiment_enabled` | `false` |
| `responsive_web_grok_temporary_chat_enabled` | `true` |
| `responsive_web_grok_text_selection_enabled` | `false` |
| `responsive_web_grok_tweet_actions_edit_image_enabled` | `false` |
| `responsive_web_grok_tweet_media_detail_edit_image_button_enabled` | `false` |
| `responsive_web_grok_tweet_media_edit_image_button_enabled` | `false` |
| `responsive_web_grok_tweet_translation` | `true` |
| `responsive_web_grok_use_new_layout` | `true` |
| `responsive_web_grok_user_active_seconds_enable` | `false` |
| `responsive_web_grok_user_seconds_debug` | `false` |
| `responsive_web_grok_v2_upsell_enabled` | `false` |
| `responsive_web_grok_voice_mode_enabled` | `false` |
| `responsive_web_grok_web_results` | `true` |
| `responsive_web_grok_webview_file_actions_enabled` | `false` |
| `responsive_web_grok_xweb_link_required` | `false` |
| `responsive_web_grok_xweb_linked_share_readonly_enabled` | `true` |
| `responsive_web_grok_xweb_replacement_enabled` | `false` |
| `responsive_web_hashtag_highlight_is_enabled` | `false` |
| `responsive_web_hashtag_highlight_show_avatar` | `false` |
| `responsive_web_hashtag_highlight_use_small_font` | `false` |
| `responsive_web_hevc_upload_preview_enabled` | `false` |
| `responsive_web_history_screen_enabled` | `true` |
| `responsive_web_home_pinned_timelines_prefetch_enabled` | `false` |
| `responsive_web_image_poll_composer_enabled` | `true` |
| `responsive_web_impression_tracker_refactor_enabled` | `true` |
| `responsive_web_in_text_shortcuts_enabled` | `true` |
| `responsive_web_install_banner_show_immediate` | `false` |
| `responsive_web_instream_video_redesign_enabled` | `true` |
| `responsive_web_intercom_support_capture_premium_enabled` | `false` |
| `responsive_web_jetfuel_add_phone_enabled` | `true` |
| `responsive_web_jetfuel_enable_sms_2fa_enabled` | `true` |
| `responsive_web_jetfuel_frame` | `true` |
| `responsive_web_lbm_v2_home_enabled` | `false` |
| `responsive_web_lbm_v2_replies_enabled` | `false` |
| `responsive_web_list_tweet_integration_enabled` | `false` |
| `responsive_web_live_commerce_enabled` | `false` |
| `responsive_web_live_scheduled_broadcast_countdown_enabled` | `false` |
| `responsive_web_live_video_viewer_session_enabled` | `true` |
| `responsive_web_livecut_native_share_enabled` | `true` |
| `responsive_web_locale_context_direction_enabled` | `true` |
| `responsive_web_location_spotlight_display_map` | `true` |
| `responsive_web_location_spotlight_v1_config` | `true` |
| `responsive_web_location_spotlight_v1_display` | `true` |
| `responsive_web_logged_out_ios_redesign_enabled` | `false` |
| `responsive_web_logged_out_redesign_enabled` | `false` |
| `responsive_web_login_input_type_email_enabled` | `false` |
| `responsive_web_login_signup_sheet_app_install_cta_enabled` | `true` |
| `responsive_web_media_download_video_share_menu_enabled` | `true` |
| `responsive_web_media_upload_appendmulti_enabled` | `true` |
| `responsive_web_media_upload_appendmulti_pre_read_blob` | `false` |
| `responsive_web_media_upload_md5_hashing_enabled` | `true` |
| `responsive_web_media_upload_metrics_enabled` | `true` |
| `responsive_web_messages_continue_enabled` | `true` |
| `responsive_web_messages_enabled` | `true` |
| `responsive_web_messages_watch_info_enabled` | `false` |
| `responsive_web_mobile_app_spotlight_v1_config` | `false` |
| `responsive_web_native_emojis_enabled` | `true` |
| `responsive_web_nested_quote_preview_enabled` | `true` |
| `responsive_web_nfl_enabled` | `true` |
| `responsive_web_nfl_game_dock_enabled` | `false` |
| `responsive_web_nfl_hub_home_tab_enabled` | `true` |
| `responsive_web_nfl_profile_sports_enabled` | `true` |
| `responsive_web_nfl_sidebar_enabled` | `true` |
| `responsive_web_nfl_sidebar_for_all_users_enabled` | `false` |
| `responsive_web_nfl_team_picker_enabled` | `true` |
| `responsive_web_not_a_bot_signups_enabled` | `false` |
| `responsive_web_ocf_reportflow_appeals_enabled` | `false` |
| `responsive_web_ocf_reportflow_dms_enabled` | `false` |
| `responsive_web_ocf_reportflow_lists_enabled` | `true` |
| `responsive_web_ocf_reportflow_profiles_enabled` | `true` |
| `responsive_web_ocf_reportflow_promoted_enabled` | `false` |
| `responsive_web_ocf_reportflow_spaces_enabled` | `false` |
| `responsive_web_ocf_reportflow_suspension_appeals_enabled` | `false` |
| `responsive_web_ocf_reportflow_testers` | `false` |
| `responsive_web_ocf_reportflow_tweets_enabled` | `true` |
| `responsive_web_ocf_sms_autoverify_darkwrite` | `false` |
| `responsive_web_ocf_sms_autoverify_enabled` | `false` |
| `responsive_web_offscreen_video_scroller_removal_enabled` | `false` |
| `responsive_web_one_hour_edit_window_enabled` | `true` |
| `responsive_web_open_in_app_prompt_enabled` | `false` |
| `responsive_web_passwordless_sso_enabled` | `false` |
| `responsive_web_personalization_id_sync_enabled` | `false` |
| `responsive_web_pinned_replies_enabled` | `false` |
| `responsive_web_placeholder_siwg_button_enabled` | `false` |
| `responsive_web_prerolls_fullscreen_disabled_on_ios` | `false` |
| `responsive_web_primary_nav_route_preload_enabled` | `false` |
| `responsive_web_priority_ntab_enabled` | `true` |
| `responsive_web_professional_journeys_holdback_enabled` | `false` |
| `responsive_web_profile_about_enabled` | `true` |
| `responsive_web_profile_redesign_enabled` | `true` |
| `responsive_web_profile_redirect_enabled` | `true` |
| `responsive_web_profile_spotlight_v0_config` | `true` |
| `responsive_web_profile_spotlight_v0_display` | `true` |
| `responsive_web_qp_ads_screening_dialog_enabled` | `false` |
| `responsive_web_qp_boost_content_check_enabled` | `true` |
| `responsive_web_qp_budgets_by_billing_currency_enabled` | `false` |
| `responsive_web_qp_full_popup_enabled` | `true` |
| `responsive_web_qp_keyword_targeting_enabled` | `false` |
| `responsive_web_qp_new_boost_analytics_enabled` | `false` |
| `responsive_web_qp_new_payment_enabled` | `false` |
| `responsive_web_qp_paused_policy_hint_enabled` | `true` |
| `responsive_web_qp_paused_rejection_reasons_enabled` | `true` |
| `responsive_web_qp_skip_objective_enabled` | `true` |
| `responsive_web_qp_two_screens_enabled` | `true` |
| `responsive_web_quick_promote_cta_enabled` | `true` |
| `responsive_web_quick_promote_high_budget_tier_enabled` | `false` |
| `responsive_web_redux_use_fragment_enabled` | `false` |
| `responsive_web_remove_qp_ad_label_enabled` | `true` |
| `responsive_web_repeat_profile_visits_notifications_device_follow_only_version_enabled` | `false` |
| `responsive_web_repeat_profile_visits_notifications_enabled` | `false` |
| `responsive_web_reply_storm_enabled` | `false` |
| `responsive_web_report_page_not_found` | `false` |
| `responsive_web_scheduling_threads_enabled` | `true` |
| `responsive_web_scribe_media_component` | `true` |
| `responsive_web_scroller_top_positioning_enabled` | `false` |
| `responsive_web_send_cookies_metadata_enabled` | `true` |
| `responsive_web_send_jetfuel_preview_image_enabled` | `true` |
| `responsive_web_share_only_tweet_url_omit_title_and_text` | `true` |
| `responsive_web_show_similar_posts_action_enabled` | `false` |
| `responsive_web_sidebar_ttf_enabled` | `false` |
| `responsive_web_signup_direct` | `false` |
| `responsive_web_sports_live_profile_rings_enabled` | `true` |
| `responsive_web_spud_enabled` | `true` |
| `responsive_web_sso_redirect_enabled` | `true` |
| `responsive_web_ssr_send_likes_in_title_enabled` | `true` |
| `responsive_web_stripe_account_creation_enabled` | `true` |
| `responsive_web_subscribers_ntab_for_creators_enabled` | `false` |
| `responsive_web_subscriptions_setting_enabled` | `true` |
| `responsive_web_suppress_app_button_banner_suppressed` | `false` |
| `responsive_web_temporary_ocf_x_migration` | `true` |
| `responsive_web_thread_media_ensure_root_urt` | `true` |
| `responsive_web_thread_media_nav_enabled` | `true` |
| `responsive_web_thread_media_tooltip` | `true` |
| `responsive_web_timeline_cover_killswitch_enabled` | `false` |
| `responsive_web_timeline_relay_lists_management_enabled` | `false` |
| `responsive_web_timeline_relay_user_lists_enabled` | `false` |
| `responsive_web_timezone_header_enabled` | `false` |
| `responsive_web_translation_feedback_enabled` | `true` |
| `responsive_web_trends_setting_new_endpoints` | `true` |
| `responsive_web_trends_ui_community_notes_enabled` | `false` |
| `responsive_web_trends_ui_enable_new_sidebar` | `false` |
| `responsive_web_trends_ui_hide_news_sidebar_on_explore` | `false` |
| `responsive_web_trends_ui_top_articles` | `true` |
| `responsive_web_tv_cast_enabled` | `true` |
| `responsive_web_tweet_analytics_m3_enabled` | `false` |
| `responsive_web_tweet_details_prefetch_enabled` | `true` |
| `responsive_web_tweet_drafts_threads_enabled` | `true` |
| `responsive_web_tweet_drafts_video_enabled` | `true` |
| `responsive_web_twitter_article_batch_posts` | `true` |
| `responsive_web_twitter_article_code_block_enabled` | `true` |
| `responsive_web_twitter_article_code_language_typeahead_enabled` | `true` |
| `responsive_web_twitter_article_latex_enabled` | `true` |
| `responsive_web_twitter_article_markdown_enabled` | `false` |
| `responsive_web_twitter_article_notes_tab_enabled` | `true` |
| `responsive_web_twitter_article_plain_text_enabled` | `true` |
| `responsive_web_twitter_article_preview_cta_redirect_enabled` | `true` |
| `responsive_web_twitter_article_reader_enabled` | `true` |
| `responsive_web_twitter_article_redirect_enabled` | `true` |
| `responsive_web_twitter_article_seed_tweet_detail_enabled` | `true` |
| `responsive_web_twitter_article_seed_tweet_enabled` | `true` |
| `responsive_web_twitter_article_table_enabled` | `true` |
| `responsive_web_twitter_article_tweet_consumption_enabled` | `true` |
| `responsive_web_twitter_blue_subscriptions_disabled` | `false` |
| `responsive_web_twitter_blue_verified_badge_ntab_empty_state_enabled` | `true` |
| `responsive_web_use_app_prompt_enabled` | `false` |
| `responsive_web_use_new_landing_page` | `true` |
| `responsive_web_use_new_onboarding` | `true` |
| `responsive_web_user_badge_education_get_verified_button_enabled` | `true` |
| `responsive_web_user_premium_user_gate` | `false` |
| `responsive_web_user_spectral_key_enabled` | `true` |
| `responsive_web_verified_ntab_hidden` | `true` |
| `responsive_web_verified_organizations_application_requirements_enabled` | `false` |
| `responsive_web_verified_organizations_application_two_column_enabled` | `false` |
| `responsive_web_verified_organizations_enterprise_insights_enabled` | `false` |
| `responsive_web_verified_organizations_enterprise_tier` | `false` |
| `responsive_web_verified_organizations_free_to_invoice_enabled` | `false` |
| `responsive_web_verified_organizations_free_upgrade_promo_enabled` | `true` |
| `responsive_web_verified_organizations_handle_form_enabled` | `false` |
| `responsive_web_verified_organizations_idv_enabled` | `false` |
| `responsive_web_verified_organizations_insights_enabled` | `true` |
| `responsive_web_verified_organizations_intercom_enabled` | `true` |
| `responsive_web_verified_organizations_invoice_enabled` | `false` |
| `responsive_web_verified_organizations_invoice_update_enabled` | `false` |
| `responsive_web_verified_organizations_new_signup_enabled` | `true` |
| `responsive_web_verified_organizations_new_year_offer_enabled` | `true` |
| `responsive_web_verified_organizations_offer_description_enabled` | `false` |
| `responsive_web_verified_organizations_paid_to_invoice_enabled` | `false` |
| `responsive_web_verified_organizations_people_search_enabled` | `false` |
| `responsive_web_verified_organizations_xbusiness_enabled` | `false` |
| `responsive_web_video_autoplay_bandwidth_threshold_enabled` | `true` |
| `responsive_web_video_pcomplete_enabled` | `true` |
| `responsive_web_video_promoted_logging_enabled` | `false` |
| `responsive_web_video_trimmer_enabled` | `false` |
| `responsive_web_vo_annual_credit_increase_enabled` | `true` |
| `responsive_web_vo_basic_application_enabled` | `true` |
| `responsive_web_x_translation_enabled` | `false` |
| `rweb_age_assurance_flow_enabled` | `true` |
| `rweb_analytics_active_followers_enabled` | `true` |
| `rweb_analytics_audience_compact_mode` | `true` |
| `rweb_analytics_audience_xweb_enabled` | `true` |
| `rweb_analytics_control_v2_enabled` | `true` |
| `rweb_analytics_export_data_content_enabled` | `true` |
| `rweb_analytics_export_data_enabled` | `true` |
| `rweb_analytics_hourly_overview_enabled` | `false` |
| `rweb_analytics_in_out_network_enabled` | `true` |
| `rweb_analytics_live_details_enabled` | `true` |
| `rweb_analytics_live_overview_enabled` | `true` |
| `rweb_analytics_nav_item_enabled` | `false` |
| `rweb_analytics_post_details_realtime_enabled` | `true` |
| `rweb_analytics_spaces_details_enabled` | `true` |
| `rweb_analytics_spaces_overview_enabled` | `true` |
| `rweb_analytics_theme` | `false` |
| `rweb_analytics_xweb_content_page` | `true` |
| `rweb_cashtags_composer_attachment_enabled` | `true` |
| `rweb_cashtags_composer_attachment_size_customization_enabled` | `true` |
| `rweb_cashtags_composer_enabled` | `true` |
| `rweb_cashtags_enabled` | `true` |
| `rweb_client_transaction_id_enabled` | `false` |
| `rweb_composer_emoji_typeahead_enabled` | `true` |
| `rweb_conf_dev_enabled` | `false` |
| `rweb_conf_multi_video_enabled` | `true` |
| `rweb_conf_only_enabled` | `false` |
| `rweb_conf_rnnoise_enabled` | `true` |
| `rweb_dash_menu_app_redirect_footer_enabled` | `true` |
| `rweb_debugger_enabled` | `false` |
| `rweb_home_jot_context_enabled` | `true` |
| `rweb_home_jot_migrate_enabled` | `true` |
| `rweb_home_mixer_enable_social_context_filter_social_contexts` | `true` |
| `rweb_home_nav_single_direction_scroll_enabled` | `false` |
| `rweb_home_ranked_following_enabled` | `true` |
| `rweb_home_uas_enabled` | `true` |
| `rweb_live_broadcast_rewind_enabled` | `true` |
| `rweb_live_dock_enabled` | `true` |
| `rweb_master_detail_enabled` | `false` |
| `rweb_media_carousel_enabled` | `true` |
| `rweb_media_multi_requests_enabled` | `true` |
| `rweb_modal_animation_enabled` | `true` |
| `rweb_modal_composer_animation_enabled` | `true` |
| `rweb_mvr_blurred_media_interstitial_enabled` | `true` |
| `rweb_navbar_grok_indicator_enabled` | `false` |
| `rweb_panning_nav_behavior` | `true` |
| `rweb_picture_in_picture_enabled` | `true` |
| `rweb_premium_business_rebranding_enabled` | `true` |
| `rweb_premium_business_rebranding_entry_point_removed` | `false` |
| `rweb_premium_business_rebranding_governments_enabled` | `true` |
| `rweb_premium_business_rebranding_hiring_url_redirect_enabled` | `true` |
| `rweb_premium_business_rebranding_landing_page_enabled` | `true` |
| `rweb_premium_business_rebranding_premium_paywall_enabled` | `true` |
| `rweb_premium_business_rebranding_premium_paywall_four_cards_enabled` | `false` |
| `rweb_premium_business_rebranding_url_enabled` | `true` |
| `rweb_quick_promote_action_menu_enabled` | `true` |
| `rweb_quick_promote_boost_enabled` | `false` |
| `rweb_quick_promote_gold_verified_boost_enabled` | `true` |
| `rweb_quick_promote_third_party_boost_enabled` | `false` |
| `rweb_recommendations_sidebar_graphql_enabled` | `true` |
| `rweb_save_video_progress_enabled` | `false` |
| `rweb_search_media_enabled` | `true` |
| `rweb_session_binding_enabled` | `false` |
| `rweb_spaces_invite_search_enabled` | `true` |
| `rweb_spaces_next_codec_enabled` | `true` |
| `rweb_sports_post_context_enabled` | `true` |
| `rweb_sports_post_context_footer_enabled` | `true` |
| `rweb_sports_post_context_header_enabled` | `false` |
| `rweb_ssp_ads_enabled` | `false` |
| `rweb_ssp_ads_premium_bypass_enabled` | `false` |
| `rweb_ssp_ads_refresh_enabled` | `false` |
| `rweb_starter_packs_topics_tab_enabled` | `false` |
| `rweb_timeline_simple_conversation_control_education_enabled` | `true` |
| `rweb_tipjar_consumption_enabled` | `false` |
| `rweb_tweets_boosting_enabled` | `false` |
| `rweb_tweets_reply_context_hidden` | `true` |
| `rweb_under_the_hood_report_enabled` | `true` |
| `rweb_under_the_hood_report_jetfuel_enabled` | `true` |
| `rweb_video_logged_in_analytics_enabled` | `true` |
| `rweb_video_pip_enabled` | `true` |
| `rweb_video_screen_enabled` | `false` |
| `rweb_video_tagging_enabled` | `true` |
| `rweb_xchat_call_health_enabled` | `true` |
| `rweb_xchat_call_link_enabled` | `false` |
| `rweb_xchat_call_redirect_enabled` | `false` |
| `rweb_xchat_call_window_enabled` | `false` |
| `rweb_xchat_calls_rtc_api_enabled` | `true` |
| `rweb_xchat_calls_v2_enabled` | `false` |
| `rweb_xchat_calls_whiteboard_enabled` | `false` |
| `rweb_xchat_client_stats_enabled` | `true` |
| `rweb_xchat_debug_enabled` | `false` |
| `rweb_xchat_dogfood_logs_enabled` | `false` |
| `rweb_xchat_logs` | `false` |
| `rweb_xchat_loudness_control_enabled` | `false` |
| `rweb_xchat_markdown_tables_enabled` | `false` |
| `rweb_xchat_scroller_rewrite_enabled` | `true` |
| `rweb_xchat_sentry_enabled` | `true` |
| `rweb_xchat_sqlite_logs` | `false` |
| `rweb_xchat_standalone_avcall_enabled` | `true` |
| `rweb_xchat_ws_only_when_active` | `false` |
| `sc_mock_data_enabled` | `false` |
| `sc_r4_enabled` | `false` |
| `search_timelines_graphql_enabled` | `true` |
| `sensitive_media_settings_enabled` | `false` |
| `sensitive_tweet_warnings_enabled` | `true` |
| `settings_for_you_recommendation_enabled` | `false` |
| `smbo_legacy_pac_is_in_follow_position_test` | `false` |
| `social_context_and_topic_context_refresh_alignment_enabled` | `false` |
| `spaces_2022_h2_clipping` | `true` |
| `spaces_2022_h2_clipping_consumption` | `true` |
| `spaces_2022_h2_spaces_communities` | `true` |
| `spaces_conference_enabled` | `false` |
| `spaces_conference_opus_dtx_enabled` | `false` |
| `spaces_dtx_opus_dtx_enabled` | `false` |
| `spaces_live_chat_enabled` | `false` |
| `spaces_video_admins_enabled` | `false` |
| `spaces_video_consumption_enabled` | `true` |
| `spaces_video_creation_enabled` | `false` |
| `spaces_video_speakers_enabled` | `false` |
| `standardized_nudges_misinfo` | `true` |
| `subscriptions_block_ad_upsell_enabled` | `false` |
| `subscriptions_blue_premium_labeling_enabled` | `true` |
| `subscriptions_blue_verified_edit_profile_error_message_enabled` | `true` |
| `subscriptions_branding_checkmark_logo_enabled` | `true` |
| `subscriptions_enabled` | `true` |
| `subscriptions_feature_1002` | `true` |
| `subscriptions_feature_1003` | `true` |
| `subscriptions_feature_1005` | `true` |
| `subscriptions_feature_1009` | `true` |
| `subscriptions_feature_1011` | `true` |
| `subscriptions_feature_1012` | `true` |
| `subscriptions_feature_1013` | `false` |
| `subscriptions_feature_1014` | `true` |
| `subscriptions_feature_account_analytics` | `true` |
| `subscriptions_feature_article_composer` | `true` |
| `subscriptions_feature_can_gift_premium` | `false` |
| `subscriptions_feature_create_premium_content` | `false` |
| `subscriptions_feature_extend_profile` | `false` |
| `subscriptions_feature_hide_subscriptions` | `true` |
| `subscriptions_feature_highlights` | `true` |
| `subscriptions_feature_labs_1004` | `true` |
| `subscriptions_feature_organization_affiliates` | `true` |
| `subscriptions_feature_organization_x_hiring` | `false` |
| `subscriptions_feature_premium_insights` | `true` |
| `subscriptions_feature_premium_jobs` | `false` |
| `subscriptions_features_premium_real_syscache_write` | `true` |
| `subscriptions_features_premium_syscache_write` | `true` |
| `subscriptions_features_syscache_read` | `true` |
| `subscriptions_features_syscache_write` | `true` |
| `subscriptions_gifting_premium_intervals_enabled` | `true` |
| `subscriptions_gifting_premium_intro_copy_enabled` | `false` |
| `subscriptions_gifting_tooltip_discount_label` | `false` |
| `subscriptions_gifting_tooltip_enabled` | `false` |
| `subscriptions_hide_ad_upsell_enabled` | `false` |
| `subscriptions_inapp_grok` | `true` |
| `subscriptions_inapp_grok_analyze` | `false` |
| `subscriptions_inapp_grok_upsell_enabled` | `true` |
| `subscriptions_is_blue_verified_review_status_profile_enabled` | `true` |
| `subscriptions_long_video_upload` | `true` |
| `subscriptions_management_billing_label_enabled` | `true` |
| `subscriptions_management_failed_payment_api_call_enabled` | `true` |
| `subscriptions_management_failed_payment_menu_alert_enabled` | `false` |
| `subscriptions_management_failed_payment_message_premium_enabled` | `false` |
| `subscriptions_management_failed_payment_paywall_banner_enabled` | `true` |
| `subscriptions_management_failed_payment_profile_card_enabled` | `false` |
| `subscriptions_management_fetch_next_billing_time` | `true` |
| `subscriptions_management_manage_subtext_update_enabled` | `false` |
| `subscriptions_management_query_active_price` | `true` |
| `subscriptions_management_renew_module_api_enabled` | `true` |
| `subscriptions_management_renew_module_enabled` | `true` |
| `subscriptions_management_tier_switch_polling_enabled` | `true` |
| `subscriptions_management_tier_switch_success_screen_enabled` | `true` |
| `subscriptions_management_use_active_price` | `true` |
| `subscriptions_marketing_page_compare_table_ctas_enabled` | `false` |
| `subscriptions_marketing_page_discounts_enabled` | `true` |
| `subscriptions_marketing_page_feature_highlights_enabled` | `false` |
| `subscriptions_marketing_page_fetch_promotions` | `true` |
| `subscriptions_marketing_page_free_trial_enabled` | `true` |
| `subscriptions_marketing_page_grok_4_web_paywall` | `false` |
| `subscriptions_marketing_page_include_tax_enabled` | `true` |
| `subscriptions_marketing_page_new_disclaimer_enabled` | `false` |
| `subscriptions_marketing_page_retention_paywall_new_button_label` | `false` |
| `subscriptions_marketing_page_social_proof_enabled` | `false` |
| `subscriptions_marketing_page_stripe_optimization_enabled` | `true` |
| `subscriptions_mute_ad_upsell_enabled` | `false` |
| `subscriptions_offers_churn_prevention_enabled` | `true` |
| `subscriptions_offers_dynamic_upsells_enabled` | `true` |
| `subscriptions_offers_in_tier_switch_enabled` | `false` |
| `subscriptions_offers_intro_title_new_copy_enabled` | `false` |
| `subscriptions_offers_localized_pricing_enabled` | `false` |
| `subscriptions_offers_paywall_urgent_heading_enabled` | `false` |
| `subscriptions_offers_premium_nav_indicator_enabled` | `false` |
| `subscriptions_offers_special_perk_enabled` | `false` |
| `subscriptions_offers_upgrade_offer_home_nav_upsell_enabled` | `false` |
| `subscriptions_offers_upgrade_offer_sidebar_upsell_enabled` | `false` |
| `subscriptions_offers_user_location_is_usa` | `false` |
| `subscriptions_premium_experiment_nav_text` | `false` |
| `subscriptions_premium_hub_ad_free_link_enabled` | `true` |
| `subscriptions_premium_hub_boost_block_enabled` | `true` |
| `subscriptions_premium_hub_insights_block_enabled` | `true` |
| `subscriptions_premium_hub_more_benefits_section_enabled` | `true` |
| `subscriptions_premium_tiers_hide_basic` | `false` |
| `subscriptions_premium_tiers_hide_basic_webview_paywall` | `false` |
| `subscriptions_quick_free_trials_low_threshold_screen_enabled` | `false` |
| `subscriptions_quick_free_trials_ui_enabled` | `false` |
| `subscriptions_report_ad_upsell_enabled` | `false` |
| `subscriptions_sign_up_enabled` | `false` |
| `subscriptions_stripe_testing` | `false` |
| `subscriptions_upsells_analytics_eligibility_query_enabled` | `false` |
| `subscriptions_upsells_analytics_fix_enabled` | `true` |
| `subscriptions_upsells_analytics_profile_enabled` | `true` |
| `subscriptions_upsells_api_enabled` | `false` |
| `subscriptions_upsells_app_tab_bar_analytics_upsell_enabled` | `false` |
| `subscriptions_upsells_articles_post_composer_promo_variant_enabled` | `false` |
| `subscriptions_upsells_articles_profile_promo_variant_enabled` | `false` |
| `subscriptions_upsells_bookmarks_screen_enabled` | `false` |
| `subscriptions_upsells_dm_card_enabled` | `false` |
| `subscriptions_upsells_edit_post_promo_variant_enabled` | `false` |
| `subscriptions_upsells_explore_sidebar_analytics_upsell_enabled` | `false` |
| `subscriptions_upsells_get_verified_button_promo_variant_enabled` | `false` |
| `subscriptions_upsells_get_verified_profile` | `true` |
| `subscriptions_upsells_get_verified_profile_card` | `true` |
| `subscriptions_upsells_get_verified_profile_card_promo_variant_enabled` | `false` |
| `subscriptions_upsells_get_verified_profile_rotation_basic_upgrade_enabled` | `true` |
| `subscriptions_upsells_get_verified_profile_rotation_enabled` | `true` |
| `subscriptions_upsells_highlights_profile_promo_variant_enabled` | `false` |
| `subscriptions_upsells_home_nav_migration_enabled` | `false` |
| `subscriptions_upsells_home_sidebar_grok_promo` | `false` |
| `subscriptions_upsells_home_sidebar_migration_enabled` | `false` |
| `subscriptions_upsells_monetization_redesign_enabled` | `true` |
| `subscriptions_upsells_post_analytics_promo_variant_enabled` | `false` |
| `subscriptions_upsells_post_details_analytics_enabled` | `true` |
| `subscriptions_upsells_post_engagements_enabled` | `false` |
| `subscriptions_upsells_post_limit_toast` | `true` |
| `subscriptions_upsells_premium_home_nav_promo_variant_enabled` | `false` |
| `subscriptions_upsells_premium_nav_migration_enabled` | `false` |
| `subscriptions_upsells_profile_card_enabled` | `false` |
| `subscriptions_upsells_profile_sidebar_analytics_upsell_enabled` | `false` |
| `subscriptions_upsells_radar_sidebar_enabled` | `false` |
| `subscriptions_upsells_reply_boost_enabled` | `false` |
| `subscriptions_upsells_reply_boost_popup_enabled` | `true` |
| `subscriptions_upsells_settings_analytics_upsell_enabled` | `false` |
| `subscriptions_upsells_sidebar_default_promo_variant_enabled` | `false` |
| `subscriptions_upsells_track_interactions_enabled` | `false` |
| `subscriptions_upsells_verified_profile_sidebar_enabled` | `false` |
| `subscriptions_upsells_verified_profile_visitor_upsell_enabled` | `true` |
| `subscriptions_upsells_visitor_get_verified_age_gate_enabled` | `true` |
| `subscriptions_upsells_vo_nav_decoration_enabled` | `false` |
| `subscriptions_verification_info_is_identity_verified_enabled` | `true` |
| `subscriptions_verification_info_verified_since_enabled` | `true` |
| `subscriptions_verified_to_premium_enabled` | `false` |
| `super_follow_android_web_subscription_enabled` | `false` |
| `super_follow_exclusive_tweet_creation_api_enabled` | `true` |
| `super_follow_onboarding_application_perks_enabled` | `true` |
| `super_follow_onboarding_granular_pricing_enabled` | `true` |
| `super_follow_subscriptions_tax_calculation_enabled` | `true` |
| `super_follow_web_application_enabled` | `false` |
| `super_follow_web_deactivate_enabled` | `true` |
| `super_follow_web_debug_enabled` | `false` |
| `super_follow_web_edit_perks_enabled` | `true` |
| `super_follow_web_onboarding_enabled` | `true` |
| `syscache_business_cancel_flow_warning_enabed` | `false` |
| `syscache_entrypoint_settings_enabled` | `true` |
| `syscache_entrypoint_vo_portal_basic_users_enabled` | `true` |
| `syscache_entrypoint_vo_portal_enabled` | `true` |
| `syscache_handle_share_banner_enabled` | `true` |
| `syscache_premium_cancel_flow_warning_enabed` | `true` |
| `syscache_syscache_pb_sidebar_handles_enabled` | `false` |
| `syscache_vo_paywall_enabled` | `true` |
| `system_theme_toggle_enabled` | `true` |
| `targeted_project_friday_enabled` | `false` |
| `timeline_scroll_pointer_events_optimization` | `false` |
| `topic_landing_page_clearer_controls_enabled` | `true` |
| `topic_landing_page_share_enabled` | `true` |
| `topics_context_controls_implicit_context_x_enabled` | `true` |
| `topics_context_controls_inline_prompt_enabled` | `false` |
| `toxic_reply_filter_inline_callout_enabled` | `false` |
| `toxic_reply_filter_settings_enabled` | `false` |
| `trusted_friends_consumption_enabled` | `true` |
| `tv_app_casting_log_focused_element_every_10s` | `false` |
| `tv_app_qrcode_login_enabled` | `true` |
| `tv_app_samsung_continue_watching_enabled` | `false` |
| `tweet_limited_actions_config_dpa_enabled` | `true` |
| `tweet_limited_actions_config_enabled` | `true` |
| `tweet_with_visibility_results_all_gql_limited_actions_enabled` | `false` |
| `tweet_with_visibility_results_partial_gql_limited_actions_enabled` | `true` |
| `tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled` | `true` |
| `twitter_chat_communities_chat_enabled` | `false` |
| `twitter_jetfuel_use_new_api_url` | `true` |
| `ucpd_enabled` | `true` |
| `unified_cards_clip_long_media_promoted_content_enabled` | `true` |
| `unified_cards_destination_url_params_enabled` | `false` |
| `unified_cards_dpa_cta_button_enabled` | `false` |
| `unified_cards_dpa_hide_vanity` | `true` |
| `unified_cards_dpa_metadata_enabled` | `true` |
| `unified_cards_hide_collection_ad_card_details` | `true` |
| `unified_cards_install_button_redesign_enabled` | `true` |
| `unified_cards_use_subtitle_as_vanity_fallback_in_collection` | `true` |
| `user_ad_accounts_config_enabled` | `false` |
| `verified_phone_label_enabled` | `false` |
| `verified_vo_refreshed_advertising_screen_enabled` | `true` |
| `video_attribution_display_over_video_cards_enabled` | `true` |
| `video_upload_metadata_title_enabled` | `false` |
| `view_counts_everywhere_api_enabled` | `true` |
| `view_counts_public_visibility_enabled` | `true` |
| `vo_upsell_enabled` | `true` |
| `vo_upsell_new_business_query_enabled` | `true` |
| `vo_upsell_profile_button_enabled` | `false` |
| `vod_attribution_tweet_detail_pivot_enabled` | `true` |
| `voice_consumption_enabled` | `true` |
| `voice_rooms_cohosts_enabled` | `true` |
| `voice_rooms_discovery_page_enabled` | `false` |
| `voice_rooms_employee_only_enabled` | `false` |
| `voice_rooms_recent_search_audiospace_ring_enabled` | `true` |
| `voice_rooms_search_results_page_audiospace_ring_enabled` | `false` |
| `voice_rooms_typeahead_audiospace_ring_enabled` | `true` |
| `voice_rooms_web_space_creation` | `true` |
| `web_video_caption_repositioning_enabled` | `true` |
| `web_video_hls_android_mse_enabled` | `true` |
| `web_video_persist_bandwidth_estimate_enabled` | `true` |
| `web_video_persist_player_preferences_enabled` | `false` |
| `web_video_playback_rate_enabled` | `true` |
| `web_video_prefetch_playlist_autoplay_disabled` | `false` |
| `web_video_safari_hlsjs_enabled` | `true` |
| `web_video_transcribed_captions_enabled` | `true` |
| `x_android_apollo_graphql_accept_header_override` | `true` |
| `x_jetfuel_enable_test_cluster` | `false` |
| `x_jetfuel_use_new_api_url` | `true` |
| `x_lite_quick_promote_analytics_banner_enabled` | `false` |
| `x_sports_game_chat_should_hide_logos` | `false` |
| `x_sports_nfl_favorite_teams_enabled` | `true` |
| `x_sports_nfl_game_chat_enabled` | `true` |
| `x_sports_nfl_game_chat_team_pick_enabled` | `true` |
| `x_sports_nfl_game_feed_enabled` | `true` |
| `x_sports_nfl_game_odds_card_enabled` | `true` |
| `x_sports_nfl_game_player_props_enabled` | `false` |
| `x_sports_nfl_game_roster_enabled` | `false` |
| `x_sports_nfl_schedule_reminder_enabled` | `true` |
| `x_sports_nfl_scorecard_odds_disabled` | `true` |
| `x_sports_play_video_enabled` | `true` |
| `x_sports_post_context_enabled` | `false` |
| `xai_profile_redirect_enabled` | `true` |
| `xcall_item_in_list_enabled` | `false` |
| `xcall_link_devices_enabled` | `false` |
| `xchat_additional_reply_preview_validation_send` | `false` |
| `xchat_ask_grok_enabled` | `true` |
| `xchat_attachment_viewer_enabled` | `false` |
| `xchat_av_call_card_interaction_enabled` | `true` |
| `xchat_av_call_start_should_notify` | `false` |
| `xchat_av_pip_enabled` | `false` |
| `xchat_batch_updates_for_bottom_cursor_processing` | `true` |
| `xchat_ckey_recovery_enabled` | `false` |
| `xchat_clear_chat_enabled` | `true` |
| `xchat_composer_photo_editor_enabled` | `false` |
| `xchat_creator_group_syncing_enabled` | `false` |
| `xchat_dm_media_upload_workmanager_enabled` | `false` |
| `xchat_dm_write_gate_enabled` | `true` |
| `xchat_drafts_bump_to_top` | `true` |
| `xchat_drop_messages_failing_franking_verification` | `false` |
| `xchat_enable_av` | `true` |
| `xchat_enable_av_group` | `true` |
| `xchat_enable_av_mobile` | `false` |
| `xchat_enable_batch_sql_events` | `false` |
| `xchat_enable_command_menu` | `true` |
| `xchat_enable_default_disappearing_messages_setting` | `false` |
| `xchat_enable_default_screenshot_blocking_setting` | `false` |
| `xchat_enable_drafts` | `false` |
| `xchat_enable_eu_report` | `false` |
| `xchat_enable_forward_message_v2` | `true` |
| `xchat_enable_franking_report` | `true` |
| `xchat_enable_in_memory_event_retry` | `true` |
| `xchat_enable_in_memory_event_retry_bottom_cursor` | `true` |
| `xchat_enable_legacy_metadata_write` | `true` |
| `xchat_enable_message_request_ui_v3` | `true` |
| `xchat_enable_message_requests` | `true` |
| `xchat_enable_message_requests_attachments` | `false` |
| `xchat_enable_metadata_sync` | `false` |
| `xchat_enable_numbers` | `true` |
| `xchat_enable_numbers_copying` | `true` |
| `xchat_enable_numbers_generation` | `false` |
| `xchat_enable_numbers_premium` | `false` |
| `xchat_enable_ratcheting` | `false` |
| `xchat_enable_reaction_notifs` | `false` |
| `xchat_enable_realm_reach_check` | `false` |
| `xchat_enable_share_message_v2` | `false` |
| `xchat_enable_single_message_locking` | `false` |
| `xchat_enable_undecryptable_tombstones` | `false` |
| `xchat_enable_user_metadata_migration_dry_run` | `false` |
| `xchat_enable_video_autoplay` | `false` |
| `xchat_enforce_media_hash_verification` | `false` |
| `xchat_external_app_integrations_enabled` | `true` |
| `xchat_fail_closed_unsigned_sends` | `true` |
| `xchat_fetch_messages_from_push_handler` | `false` |
| `xchat_grok_bots` | `false` |
| `xchat_group_admin_settings_enabled` | `true` |
| `xchat_group_description_enabled` | `true` |
| `xchat_grpc_chat_page_enabled` | `false` |
| `xchat_hide_pending_members_for_non_admin` | `true` |
| `xchat_inbox_checksum_scribing_enabled` | `true` |
| `xchat_inbox_new_messages_preview_enabled` | `false` |
| `xchat_invoke_grok_bot_enabled` | `false` |
| `xchat_keep_db_when_processing_inbox_response` | `false` |
| `xchat_keypair_nuclear_reset_enabled` | `false` |
| `xchat_keypair_recovery_enabled` | `false` |
| `xchat_liquid_glass_convo_header_enabled` | `false` |
| `xchat_long_messages_enabled` | `true` |
| `xchat_markdown_rendering_enabled` | `true` |
| `xchat_media_download_io_batching_enabled` | `true` |
| `xchat_media_upload_io_batching_enabled` | `true` |
| `xchat_message_composer_v2` | `false` |
| `xchat_message_date_tap_enabled` | `false` |
| `xchat_message_list_v2_enabled` | `false` |
| `xchat_message_request_rate_limit_upsell` | `true` |
| `xchat_missing_messages_scribe_enabled` | `false` |
| `xchat_molecule_enabled` | `false` |
| `xchat_observe_inbox_categories_separately` | `true` |
| `xchat_observe_inbox_users_enabled` | `true` |
| `xchat_pay_button_enabled` | `true` |
| `xchat_pin_signatures_enabled` | `false` |
| `xchat_pinned_messages_enabled` | `true` |
| `xchat_pinned_messages_reading_enabled` | `true` |
| `xchat_plaintext_grok_enabled` | `false` |
| `xchat_pmk_bundle_signatures_enabled` | `false` |
| `xchat_realm_migration_local_key_fallback_enabled` | `false` |
| `xchat_repair_groups_after_passcode_reset` | `true` |
| `xchat_resolve_unencrypted_media_locally` | `true` |
| `xchat_self_custody_enabled` | `false` |
| `xchat_send_franking_data` | `true` |
| `xchat_separate_reactions_table_enabled` | `false` |
| `xchat_settings_enabled` | `false` |
| `xchat_settings_redesign_enabled` | `true` |
| `xchat_share_to_ig_story` | `false` |
| `xchat_standalone_push_notifications` | `false` |
| `xchat_streamed_link_preview` | `true` |
| `xchat_swipe_to_reveal_timestamp_enabled` | `false` |
| `xchat_throttle_badge_counts` | `false` |
| `xchat_unread_mismatch_scribing_enabled` | `true` |
| `xchat_url_preview_v2_enabled` | `true` |
| `xchat_validate_priv_pub_on_load` | `false` |
| `xchat_voice_messages_enabled` | `true` |
| `xchat_voice_messages_transcription_enabled` | `false` |
| `xprofile_consumption_enabled` | `false` |
| `xprofile_editing_enabled` | `false` |
| `xprofile_emojis_enabled` | `true` |
| `xprofile_profile_button_enabled` | `false` |
| `xprofile_section_visibility_enabled` | `false` |
| `xprofile_work_history_consumption_enabled` | `true` |
| `xprofile_work_history_domain_enabled` | `false` |
| `xprofile_work_history_enabled` | `true` |
| `xstudio_rweb_live_studio_enabled` | `true` |

## Valeurs de repli intégrées à la CLI

Ces valeurs servent lorsque la correspondance propre à une opération est absente. Elles proviennent de `core/constants.go`, indépendamment du relevé dynamique.

| Indicateur | Repli |
| --- | --- |
| `articles_preview_enabled` | `true` |
| `c9s_tweet_anatomy_moderator_badge_enabled` | `true` |
| `communities_web_enable_tweet_community_results_fetch` | `true` |
| `creator_subscriptions_quote_tweet_preview_enabled` | `false` |
| `creator_subscriptions_tweet_preview_api_enabled` | `true` |
| `freedom_of_speech_not_reach_fetch_enabled` | `true` |
| `graphql_is_translatable_rweb_tweet_is_translatable_enabled` | `true` |
| `longform_notetweets_consumption_enabled` | `true` |
| `longform_notetweets_inline_media_enabled` | `true` |
| `longform_notetweets_rich_text_read_enabled` | `true` |
| `responsive_web_edit_tweet_api_enabled` | `true` |
| `responsive_web_enhance_cards_enabled` | `true` |
| `responsive_web_graphql_exclude_directive_enabled` | `true` |
| `responsive_web_graphql_skip_user_profile_image_extensions_enabled` | `false` |
| `responsive_web_graphql_timeline_navigation_enabled` | `true` |
| `responsive_web_twitter_article_tweet_consumption_enabled` | `true` |
| `rweb_tipjar_consumption_enabled` | `true` |
| `rweb_video_timestamps_enabled` | `true` |
| `standardized_nudges_misinfo` | `true` |
| `tweet_awards_web_tipping_enabled` | `false` |
| `tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled` | `true` |
| `verified_phone_label_enabled` | `false` |
| `view_counts_everywhere_api_enabled` | `true` |

Le basculeur de champ intégré `withArticlePlainText` vaut `false` (`DefaultFieldToggles`).

## Commandes utiles

```sh
xsh endpoints list --json
xsh endpoints status --json
xsh endpoints refresh --account default --json
xsh endpoints check UserByScreenName --account default --json
xsh endpoints check SearchTimeline --account default --json
xsh endpoints check Followers --account default --json
xsh status --check --account default --json
```

La commande `refresh` actualise l’inventaire ; `check` effectue une requête réelle uniquement pour les opérations ayant une sonde sûre. Une réponse `unsupported` signifie « aucune sonde définie », pas « endpoint opérationnel ». Pour actualiser ce document après une nouvelle découverte, recalculer les inventaires depuis le cache et vérifier les états observés.
