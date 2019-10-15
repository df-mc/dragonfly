// package endpoints implements a number of HTTP endpoints that may be called to get information about the
// server or to trigger an action, such as kicking a player.
//
// Available endpoints
// * GET http://address/player_count -> Returns a JSON object containing the player count of the server.
// * GET http://address/max_player_count -> Returns a JSON object containing the maximum amount of players.
// * GET http://address/players -> Returns a JSON object containing a list of all online players of the server and information identifying them.
// * GET http://address/mem -> Returns a JSON object containing memory usage information about the server.
// * POST http://address/kick -> Kicks a player from the server.
//   - Form values:
//     * uuid -> The UUID of the player to kick.
//     * message -> The message to kick the player with. If left empty, the player will be directed to the main menu immediately.
package endpoints
