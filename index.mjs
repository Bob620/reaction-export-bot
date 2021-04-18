import readline from 'readline';
import {promises as fs} from 'fs';

import Discord from 'discord.js';

const client = new Discord.Client();

import config from './config/config.json';

const face = readline.createInterface({
	input: process.stdin,
	output: process.stdout
});

function question(query) {
	return new Promise(resolve => face.question(query, resolve));
}

let state = {
	client,
	guild: undefined,
	channel: undefined
};

async function genInfo() {
	let info = '';

	if (state.channel)
		switch(state.channel.type) {
			case 'dm':
				info = `DM-${state.channel.id}`;
				break;
			case 'text':
			case 'news':
				info = `${state.channel.guild.name}#${state.channel.name}`;
				break;
		}
	else if (state.guild && state.guild.available)
		info = `${state.guild.name}`;
	return info;
}

async function guildMenu() {
	console.log('');
	console.log(`Enter option number: `);
	console.log(`0) Back`);
	let channels = [];
	let i = 1;
	for (const [, channel] of state.guild.channels.cache)
		if (channel.type === 'text') {
			console.log(`${i}) #${channel.name}`);
			channels[i] = channel;
			i++;
		}

	const req = await question(`[${await genInfo()}]> `);
	if (req !== '0' && channels[req])
		state.channel = channels[req];
}

async function mainMenu() {
	console.log('');
	console.log(`Enter option number: `);
	console.log(`0) Exit`);
	console.log('1) Change Guild');
	console.log('2) Search Guild Channels');
	console.log('3) Enter Channel ID');
	console.log(`4) Enter Message ID(s)`);

	const req = await question(`[${await genInfo()}]> `);
	switch(req) {
		case '1':
			state.guild = await getGuild(await question(`ID: `));
			state.channel = undefined;
			break;
		case '2':
			if (state.guild && state.guild.available)
				await guildMenu();
			else
				console.log('No channels available, try later or another option');
			break;
		case '3':
			state.channel = await getChannel(await question(`ID: `));
			break;
		case '4':
			let reactions = new Map();

			let ids = (await question(`ID( ID ID...): `)).split(' ').filter(i => i);

			let reactionPos = 0;
			for (const id of ids) {
				for (const [userId, {emote, user}] of (await processMessage(await getMessage(id)))) {
					if (!reactions.has(userId))
						reactions.set(userId, {user, emotes: []});
					reactions.get(userId).emotes[reactionPos] = emote;
				}
				reactionPos++;
			}

			await outputCSV(config.outputFile, reactions.values());
			break;
		case '0':
			process.exit();
	}

	await mainMenu();
}

async function getGuild(guildID) {
	console.log('Searching for guild...');

	const guild = state.client.guilds.cache.get(guildID);
	if (guild === undefined || !guild.available)
		throw `Unable to find guild (${guildID})`;
	console.log(`Found guild (${guildID})`);
	return guild;
}

async function getChannel(channelID) {
	console.log('Searching for channel...');

	const channel = state.client.channels.cache.get(channelID);
	if (channel === undefined || !channel.available)
		throw `Unable to find channel (${channelID})`;
	switch(channel.type) {
		case 'dm':
		case 'text':
		case 'news':
			console.log(`Found channel (${channelID}) of type ${channel.type}`);
			return channel;
		default:
			throw `Found channel but invalid type (DM, Text, or News required)`;
	}
}

async function getMessage(messageID) {
	console.log('Searching for message...');

	let message;
	if (state.channel && state.channel.available)
		try {
			message = await state.channel.messages.fetch(messageID);
		} catch(err) {
		}
	if ((!message || !message.available) && (state.guild && state.guild.available))
		for (const [, channel] of state.guild.channels.cache)
			if (channel.type !== 'voice' || 'category' || 'store')
				try {
					await channel.fetch();
					message = await channel.messages.fetch(messageID);
				} catch(err) {
				}

	return message;
}

async function processMessage(message) {
	console.log(`Processing message reactions (${message.id})...`);
	const reactions = Array.from(await message.reactions.cache.entries()).slice(0, 2);

	if (reactions.length < 2)
		console.warn('Found less than 2 types of reactions');

	let alreadyExist = [];
	let users = new Map();
	let reactionUsers;
	for (const [, reaction] of reactions) {
		let id = '';
		reactionUsers = await reaction.users.fetch();
		while (reactionUsers.size > 0) {
			let user;
			for ([id, user] of reactionUsers) {
				if (users.has(id) && !alreadyExist.includes(id))
					alreadyExist.push(id);
				users.set(id, {emote: reaction.emoji, user});
			}
			reactionUsers = await reaction.users.fetch({after: id});
		}
	}

	console.log(`${alreadyExist.length} user${alreadyExist.length === 1 ? '' : 's'} reacted two or more times ${alreadyExist.length !== 0 ? '(Tch)' : ''}`);
	return users;
}

async function outputCSV(fileName, reactions) {
	console.log('Writing out file...');
	await fs.writeFile(fileName, 'name,emote\n', {encoding: 'utf8'});
	for (const {emotes, user} of reactions) {
		let userData = `${user.tag}`;

		for (let i = 0; i < emotes.length; i++)
			if (emotes[i] !== undefined)
				userData += `,${emotes[i].name}`;
			else
				userData += ',';

		await fs.appendFile(fileName, `${userData}\n`, {encoding: 'utf8'});
	}
}

client.on('ready', async () => {
	console.log(`Logged in as ${state.client.user.tag}\n`);
	try {
		state.guild = await getGuild(config.guildID);
		state.channel = undefined;
	} catch(err) {
		console.error(err);
	}

	await mainMenu();
	face.close();
});

client.login(config.botToken);