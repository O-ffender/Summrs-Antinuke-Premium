package database

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (db *MongoDB) CreateGuild(user *discordgo.User, guild *discordgo.Guild) {
	_, err = db.FindData(guild.ID)
	if err == nil {
		return
	}

	defaultGuildData := bson.M{
		"Anti-Invite":                 	"off",
		"Anti-Ban":                     true,
		"Anti-Bot":                     true,
		"Anti-Kick":                    true,
		"Anti-Guild-Name":             	true,
		"Anti-Widget-Spam":             true,
		"Anti-Admin-Role":             	true,
		"Anti-Role-Create":             true,
		"Anti-Role-Delete":             true,
		"Anti-Vanity-Steal":            true,
		"Anti-Channel-Create":          true,
		"Anti-Channel-Delete":          true,
		"Anti-Webhook-Create":          true,
		"guild_id":                     guild.ID,
		"guild-name":                   guild.Name,
		"log-channel":                  "nil",
		"Anti-Nuke-Action":             "ban",
		"prefix":                       ">",
		"users":                        bson.A{guild.OwnerID},
		"vanity-url":                   guild.VanityURLCode,
		"whitelisted-webhook-channels": bson.A{},
		"whitelisted-invite-channels":  bson.A{},
		"whitelisted-roles":            bson.A{},
	}

	if _, err = db.Collection.InsertOne(context.Background(), defaultGuildData); err != nil {
		return
	}

	if _, err = db.Collection.UpdateOne(context.Background(), bson.M{"guild_id": guild.ID}, bson.M{"$push": bson.M{"users": user.ID}}); err != nil {
		return
	}

}

func (db *MongoDB) DeleteGuild(guildID string) bool {
	if _, err = db.Collection.DeleteOne(context.Background(), bson.M{"guild_id": guildID}); err != nil {
		return false
	}
	return true
}

func (db *MongoDB) FindData(guildID string) (bson.M, error) {
	var data bson.M
	if err = db.Collection.FindOne(context.Background(), bson.M{"guild_id": guildID}).Decode(&data); err != nil {
		return nil, err
	}

	return data, nil
}

func (db *MongoDB) IsWhitelisted(guildID, typeKey, ID string, optionalMember *discordgo.Member) bool {
	var (
		data bson.M
		err  error
	)

	data, err = db.FindData(guildID)
	if err != nil {
		return false
	}

	for _, whitelistedID := range data[typeKey].(bson.A) {
		if whitelistedID.(string) == ID {
			return true
		}
	}

	if optionalMember == nil {
		return false
	}

	for _, whitelistedRoleID := range data["whitelisted-roles"].(bson.A) {
		for _, roleID := range optionalMember.Roles {
			if whitelistedRoleID.(string) == roleID {
				return true
			}
		}
	}

	return false
}

//"$set"
func (db *MongoDB) SetData(typeKey string, guildID string, index string, value string) (bool, error) {
	if _, err = db.Collection.UpdateOne(context.Background(), bson.M{"guild_id": guildID}, bson.M{typeKey: bson.M{index: value}}, &options.UpdateOptions{}); err != nil {
		return false, err
	}
	return true, nil
}

func (db *MongoDB) SetToggle(typeKey string, guildID string, index string, value bool) (bool, error) {
	if _, err = db.Collection.UpdateOne(context.Background(), bson.M{"guild_id": guildID}, bson.M{typeKey: bson.M{index: value}}, &options.UpdateOptions{}); err != nil {
		return false, err
	}
	return true, nil
}

//"$push" = index
//"$pull" = index
//"users" = value
//"whitelisted-roles" = value

/*
	if _, err = db.Collection.UpdateOne(context.Background(), bson.M{"guild_id": guildID}, bson.M{"$pull": bson.M{"users": id}}, &options.UpdateOptions{}); err != nil {
		return false, err
	}
*/

func (db *MongoDB) SetWhitelistData(guildID, ID, index, valueKey string) (bool, error) {
	whitelisted := db.IsWhitelisted(guildID, valueKey, ID, nil)

	if index == "$push" && whitelisted {
		return false, fmt.Errorf("I couldn't seem to change their whitelist status, Maybe check whitelists?")
	}

	if _, err = db.Collection.UpdateOne(context.Background(), bson.M{"guild_id": guildID}, bson.M{index: bson.M{valueKey: ID}}, &options.UpdateOptions{}); err != nil {
		return false, err
	}

	return true, nil
}

func SetupDB() MongoDB {
	var db = MongoDB{}

	db.Client, err = mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017/summrs?retryWrites=true&w=majority"))

	if err != nil {
		panic(err)
	}

	err = db.Client.Connect(db.Ctx)
	if err != nil {
		panic(err)
	}

	db.Database = db.Client.Database("summrs")
	db.Collection = db.Database.Collection("whitelist")

	return db
}

var (
	cancel   func()
	Database = SetupDB()
	err      error
)

type (
	MongoDB struct {
		Collection *mongo.Collection
		Client     *mongo.Client
		Ctx        context.Context
		Database   *mongo.Database
	}
)
