# Politique de confidentialité

> [!NOTE]
> This privacy policy is in French, because I'm French. Feel free to translate this into your own language.

Le robot Discord nommé "Les Copaings Bot" ("bot") suit cette politique de confidentialité.

## Informations que nous récupérons

Le bot récupère les informations suivantes de tous les membres des serveurs sur lesquels il est :
- nom d'utilisateur (e.g., Anhgelus Morhtuuzh) et nom d'affichage (e.g., anhgelus)
- identifiant (e.g., 394089252733976576)
- tous les messages envoyés à partir de la date d'arrivée du bot sur le serveur Discord
- l'activité vocale

## Traitement des données 

Le nom d'utilisateur et le nom d'affichage ne sont pas sauvegardés. Ils ne servent qu'à donner du contexte lors de l'utilisation des fonctionnalités du bot. Ces informations sont nécessaires pour le bon fonctionnement du bot.

L'identifiant est sauvegardé sans hashage dans une base de donnée. Cet identifiant permet d'identifier un utilisateur pour lui attribuer certaines valeures nécessaires au bon fonctionnement du bot (comme son expérience).

Les messages envoyés ne sont pas sauvegardés. Ils sont utilisés par calculer l'expérience gagné à l'aide de ce message. Pour ce faire, le bot récupère le nombre de caractère différent ainsi que la taille du message. Ensuite, ils calculent l'expérience à l'aide de ces informations. Uniquement l'expérience est sauvegardée. Ces données sont nécessaires au bon fonctionnement du bot.

L'heure de connexion à un salon vocal est sauvegardé sur une base de donnée non persistente. Cela permet de calculer l'expérience gagné dans un salon vocal puisque ce calcul demande le temps passé en étant connecté. Après la déconnexion de l'utilisateur, les informations sont supprimées. Ces données sont nécessaires au bon fonctionnement du bot pour calculer l'expérience.

## Rétention des données et vos droits

Ces données sont sauvegardées aussi longtemps que nécessaire. Dès que l'utilisateur envoie un message ou se connecte à un salon vocal, alors des données sont sauvegardées. Elles sont automatiquement supprimées quand celui-ci quitte le serveur Discord ou que le bot est enlevé du serveur.

En tant qu'utilisateur, vous avez le droit de :
- supprimer en partie ou intégralement les données
- avoir accès à vos données sauvegardées
- modifier ces données

Pour exercer vos droits concernant l'accès aux données ou à la modification des données, contactez me@anhgelus.world.

La suppression des données entraîne de facto une remise à zéro de votre profile sur un serveur, contactez le propriétaire du serveur pour exercer ce droit.

Concernant la suppression intégrale des données, contacter me@anhgelus.world.

Si vous souhaitez que vos données ne soient plus utilisées par le bot, contacter le propriétaire du serveur ou quitter le serveur.

